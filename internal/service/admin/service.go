package admin

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"

	"privacy-social-backend/internal/repository"
	"privacy-social-backend/internal/repository/db"
)

const (
	statsCacheKey = "admin:stats"
	statsCacheTTL = 1 * time.Minute
)

type ListUsersParams struct {
	PageID   int32
	PageSize int32
}

type BanUserParams struct {
	UserID string
	Ban    bool
}

type Service interface {
	GetStats(ctx context.Context) (map[string]interface{}, bool, error) // Returns data, isCached, error
	ListUsers(ctx context.Context, params ListUsersParams) ([]db.User, int64, error)
	BanUser(ctx context.Context, params BanUserParams) (db.User, error)
	DeleteUser(ctx context.Context, userID string) error
	ListReports(ctx context.Context, resolved bool, pageID, pageSize int32) ([]db.ListReportsRow, error)
	ResolveReport(ctx context.Context, reportID string) (db.Report, error)
	DeleteStory(ctx context.Context, storyID string) error
	ListAllStories(ctx context.Context, pageID, pageSize int32) ([]db.ListAllStoriesRow, error)
}

type ServiceImpl struct {
	store repository.Store
	redis *redis.Client
}

func NewService(store repository.Store, redis *redis.Client) Service {
	return &ServiceImpl{
		store: store,
		redis: redis,
	}
}

func (s *ServiceImpl) GetStats(ctx context.Context) (map[string]interface{}, bool, error) {
	// Try cache first
	cachedData, err := s.redis.Get(ctx, statsCacheKey).Result()
	if err == nil && cachedData != "" {
		var response map[string]interface{}
		if err := json.Unmarshal([]byte(cachedData), &response); err == nil {
			return response, true, nil
		}
	}

	// Cache miss - query database
	userStats, err := s.store.GetSystemStats(ctx)
	if err != nil {
		return nil, false, err
	}

	storyStats, err := s.store.GetStoryStats(ctx)
	if err != nil {
		return nil, false, err
	}

	// Fetch Analytics (North Star)
	retention, err := s.store.GetStreakRetentionStats(ctx)
	if err != nil {
		log.Error().Err(err).Msg("failed to get retention stats")
		retention = db.GetStreakRetentionStatsRow{}
	}
	engagement, err := s.store.GetEngagementStats(ctx)
	if err != nil {
		log.Error().Err(err).Msg("failed to get engagement stats")
		engagement = db.GetEngagementStatsRow{}
	}
	conversion, err := s.store.GetConversionStats(ctx)
	if err != nil {
		log.Error().Err(err).Msg("failed to get conversion stats")
		conversion = db.GetConversionStatsRow{}
	}

	response := map[string]interface{}{
		"users":   userStats,
		"stories": storyStats,
		"analytics": map[string]interface{}{
			"retention_rate_3d":        retention.RetentionRate,
			"retained_users_count":     retention.RetainedUsersCount,
			"weekly_stories_per_user":  engagement.AvgStoriesPerUserWeekly,
			"crossing_conversion_rate": conversion.CrossingConversionRate,
		},
	}

	// Cache for 1 minute
	responseJSON, _ := json.Marshal(response)
	s.redis.Set(ctx, statsCacheKey, responseJSON, statsCacheTTL)

	return response, false, nil
}

func (s *ServiceImpl) ListUsers(ctx context.Context, params ListUsersParams) ([]db.User, int64, error) {
	users, err := s.store.ListUsers(ctx, db.ListUsersParams{
		Limit:  params.PageSize,
		Offset: (params.PageID - 1) * params.PageSize,
	})
	if err != nil {
		return nil, 0, err
	}

	count, err := s.store.CountUsers(ctx)
	if err != nil {
		return nil, 0, err
	}

	return users, count, nil
}

func (s *ServiceImpl) BanUser(ctx context.Context, params BanUserParams) (db.User, error) {
	userID, err := uuid.Parse(params.UserID)
	if err != nil {
		return db.User{}, err
	}

	return s.store.BanUser(ctx, db.BanUserParams{
		ID:             userID,
		IsShadowBanned: params.Ban,
	})
}

func (s *ServiceImpl) DeleteUser(ctx context.Context, userID string) error {
	id, err := uuid.Parse(userID)
	if err != nil {
		return err
	}
	return s.store.DeleteUser(ctx, id)
}

func (s *ServiceImpl) ListReports(ctx context.Context, resolved bool, pageID, pageSize int32) ([]db.ListReportsRow, error) {
	return s.store.ListReports(ctx, db.ListReportsParams{
		IsResolved: resolved,
		Limit:      pageSize,
		Offset:     (pageID - 1) * pageSize,
	})
}

func (s *ServiceImpl) ResolveReport(ctx context.Context, reportID string) (db.Report, error) {
	id, err := uuid.Parse(reportID)
	if err != nil {
		return db.Report{}, err
	}
	return s.store.ResolveReport(ctx, id)
}

func (s *ServiceImpl) DeleteStory(ctx context.Context, storyID string) error {
	id, err := uuid.Parse(storyID)
	if err != nil {
		return err
	}
	err = s.store.DeleteStory(ctx, id)
	if err != nil {
		return err
	}

	// Invalidate feed cache
	keys, err := s.redis.Keys(ctx, "feed:*").Result()
	if err == nil && len(keys) > 0 {
		s.redis.Del(ctx, keys...)
	}
	return nil
}

func (s *ServiceImpl) ListAllStories(ctx context.Context, pageID, pageSize int32) ([]db.ListAllStoriesRow, error) {
	return s.store.ListAllStories(ctx, db.ListAllStoriesParams{
		Limit:  pageSize,
		Offset: (pageID - 1) * pageSize,
	})
}
