package location

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/mmcloughlin/geohash"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"

	"privacy-social-backend/internal/repository"
	"privacy-social-backend/internal/repository/db"
)

const (
	// Key prefix for user locations
	// Type: GEO (Sorted Set)
	// Member: UserID
	userLocationsKey = "users:locations"

	// Key prefix for daily crossing duplications
	// Type: String (with TTL)
	// Key: crossing:<uid1>:<uid2>
	crossingKeyPrefix = "crossing:"

	// Crossing TTL (Don't trigger same crossing for 24h)
	crossingTTL = 24 * time.Hour

	// Radius for "crossing paths" (approx 76m to match Geohash precision)
	crossingRadiusMeters = 80.0
)

type RedisLocationService struct {
	redis *redis.Client
	store repository.Store
}

func NewRedisLocationService(redis *redis.Client, store repository.Store) *RedisLocationService {
	return &RedisLocationService{
		redis: redis,
		store: store,
	}
}

// UpdateUserLocation updates user position in Redis and triggers real-time crossing detection
func (s *RedisLocationService) UpdateUserLocation(ctx context.Context, userID uuid.UUID, lat, lng float64) error {
	// 1. Update Geo Index
	err := s.redis.GeoAdd(ctx, userLocationsKey, &redis.GeoLocation{
		Name:      userID.String(),
		Longitude: lng,
		Latitude:  lat,
	}).Err()
	if err != nil {
		return fmt.Errorf("failed to update geo location: %w", err)
	}

	// 2. Find nearby users (Real-time Crossing Detection)
	// look for users within specific radius
	matches, err := s.redis.GeoRadius(ctx, userLocationsKey, lng, lat, &redis.GeoRadiusQuery{
		Radius:    crossingRadiusMeters,
		Unit:      "m",
		WithDist:  true,
		WithCoord: true,
	}).Result()
	if err != nil {
		// Log but don't fail the request, basic location update succeeded
		log.Error().Err(err).Msg("failed to query nearby users")
		return nil
	}

	// 3. Process matches
	s.processCrossings(ctx, userID, matches)

	return nil
}

func (s *RedisLocationService) processCrossings(ctx context.Context, userID uuid.UUID, matches []redis.GeoLocation) {
	for _, match := range matches {
		targetUserIDStr := match.Name

		// Skip self
		if targetUserIDStr == userID.String() {
			continue
		}

		targetUserID, err := uuid.Parse(targetUserIDStr)
		if err != nil {
			continue
		}

		// Ensure consistent ordering for key generation (u1 < u2)
		u1, u2 := userID, targetUserID
		if u1.String() > u2.String() {
			u1, u2 = u2, u1
		}

		// 4. Check De-duplication Cache (Redis)
		dedupKey := fmt.Sprintf("%s%s:%s", crossingKeyPrefix, u1.String(), u2.String())
		exists, err := s.redis.Exists(ctx, dedupKey).Result()
		if err == nil && exists > 0 {
			// Already crossed recently, skip
			continue
		}

		// 5. Verify Database Constraints (Block, Ghost Mode, etc) - Optional optimization:
		// We could do this BEFORE DB insert, but `FindPotentialCrossings` query was complex.
		// For now, let's assume if they are both "active" (pinging), they are candidates.
		// We'll let the user decide if we need deep validation queries here or if we just insert.
		// Detailed validation is safer.

		// Check privacy/ghost mode for BOTH users
		// We can reuse the extensive WHERE clause from the original SQL if we want,
		// but checking individual user status is faster.

		valid, err := s.validateCrossingPrivacy(ctx, userID, targetUserID)
		if err != nil || !valid {
			continue
		}

		// 6. Record Crossing in DB
		// Calculate center geohash
		// Note: match.Longitude/Latitude might be empty if we didn't ask WithCoord, but we did.
		centerHash := geohash.Encode(match.Latitude, match.Longitude)

		crossing, err := s.store.CreateCrossing(ctx, db.CreateCrossingParams{
			UserID1:        u1,
			UserID2:        u2,
			LocationCenter: centerHash,
			OccurredAt:     time.Now().UTC(),
		})
		if err != nil {
			log.Error().Err(err).Msg("failed to persist crossing")
			continue
		}

		// 7. Send Notification
		// Only notify the OTHER person "You crossed with current_user"
		// Actually, usually both get notified or it is symmetric.
		// Original logic notified User2 about User1.

		// Notify User 1
		s.createNotification(ctx, userID, targetUserID, crossing.ID)
		// Notify User 2
		s.createNotification(ctx, targetUserID, userID, crossing.ID)

		// 8. Invalidate crossings cache for both users
		s.invalidateCrossingsCache(ctx, userID)
		s.invalidateCrossingsCache(ctx, targetUserID)

		// 9. Set Dedup Key
		s.redis.Set(ctx, dedupKey, "1", crossingTTL)
	}
}

func (s *RedisLocationService) validateCrossingPrivacy(ctx context.Context, u1, u2 uuid.UUID) (bool, error) {
	// Check blocks
	blocked, err := s.store.IsUserBlocked(ctx, db.IsUserBlockedParams{
		BlockerID: u1,
		BlockedID: u2,
	})
	if err != nil {
		return false, err
	}
	if blocked {
		return false, nil
	}

	blockedReverse, err := s.store.IsUserBlocked(ctx, db.IsUserBlockedParams{
		BlockerID: u2,
		BlockedID: u1,
	})
	if err != nil {
		return false, err
	}
	if blockedReverse {
		return false, nil
	}

	// Check Ghost Mode (using User model)
	// Ideally we cache this "IsActive" status.
	// For now, fetching user is safest.
	user1, err := s.store.GetUserByID(ctx, u1)
	if err != nil {
		return false, err
	}
	if user1.IsGhostMode || user1.IsShadowBanned {
		return false, nil
	}

	user2, err := s.store.GetUserByID(ctx, u2)
	if err != nil {
		return false, err
	}
	if user2.IsGhostMode || user2.IsShadowBanned {
		return false, nil
	}

	return true, nil
}

func (s *RedisLocationService) createNotification(ctx context.Context, recipient, crossedWith uuid.UUID, crossingID uuid.UUID) {
	_, err := s.store.CreateNotification(ctx, db.CreateNotificationParams{
		UserID:            recipient,
		Type:              "crossing_detected",
		Title:             "Path Crossed!",
		Message:           "You crossed paths with someone nearby",
		RelatedUserID:     uuid.NullUUID{UUID: crossedWith, Valid: true},
		RelatedCrossingID: uuid.NullUUID{UUID: crossingID, Valid: true},
	})
	if err != nil {
		log.Error().Err(err).Msg("failed to create notification for crossing")
	}
	// Note: We could trigger WebSocket here if we had access to hub
}

// invalidateCrossingsCache removes the cached crossings for a user
func (s *RedisLocationService) invalidateCrossingsCache(ctx context.Context, userID uuid.UUID) {
	cacheKey := "crossings:v3:" + userID.String()
	s.redis.Del(ctx, cacheKey)
}
