package safety

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	// MaxSpeedKmH is 1000 km/h (approx jet speed). Anything faster is definitely fake.
	MaxSpeedKmH = 1000.0
	// Key prefix for last location
	lastLocationKeyPrefix = "safety:last_loc:"
)

type ValidationResult struct {
	Allowed   bool
	Reason    string
	ShouldBan bool
}

// Monitor handles safety checks like Fake GPS
type Monitor struct {
	redis *redis.Client
}

func NewMonitor(rdb *redis.Client) *Monitor {
	return &Monitor{redis: rdb}
}

// ValidateUserMovement checks if the user moved typically fast
func (s *Monitor) ValidateUserMovement(ctx context.Context, userID string, newLat, newLng float64) ValidationResult {
	key := lastLocationKeyPrefix + userID

	// Get last location
	res, err := s.redis.HGetAll(ctx, key).Result()
	if err != nil || len(res) == 0 {
		// First ping or expired, just save new location
		s.saveLastLocation(ctx, key, newLat, newLng)
		return ValidationResult{Allowed: true}
	}

	// Parse last location
	lastLat := parseFloat(res["lat"])
	lastLng := parseFloat(res["lng"])
	lastTime, _ := time.Parse(time.RFC3339, res["time"])

	now := time.Now()

	// Calculate distance (Haversine)
	distKm := haversineKm(lastLat, lastLng, newLat, newLng)
	timeDiffHours := now.Sub(lastTime).Hours()

	if timeDiffHours <= 0 {
		// Same timestamp or clock skew? Allow but update
		s.saveLastLocation(ctx, key, newLat, newLng)
		return ValidationResult{Allowed: true}
	}

	speed := distKm / timeDiffHours

	if speed > MaxSpeedKmH {
		return ValidationResult{
			Allowed:   false,
			Reason:    "Speed limit exceeded (" + formatFloat(speed) + " km/h)",
			ShouldBan: true,
		}
	}

	// Valid, update last location
	s.saveLastLocation(ctx, key, newLat, newLng)
	return ValidationResult{Allowed: true}
}

func (s *Monitor) saveLastLocation(ctx context.Context, key string, lat, lng float64) {
	s.redis.HSet(ctx, key, map[string]interface{}{
		"lat":  lat,
		"lng":  lng,
		"time": time.Now().Format(time.RFC3339),
	})
	s.redis.Expire(ctx, key, 24*time.Hour)
}

// -- Helpers --

func haversineKm(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371 // Radius of the earth in km
	dLat := (lat2 - lat1) * (math.Pi / 180.0)
	dLon := (lon2 - lon1) * (math.Pi / 180.0)
	lat1 = lat1 * (math.Pi / 180.0)
	lat2 = lat2 * (math.Pi / 180.0)

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Sin(dLon/2)*math.Sin(dLon/2)*math.Cos(lat1)*math.Cos(lat2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return R * c
}

func parseFloat(s string) float64 {
	f, _ := strconv.ParseFloat(s, 64)
	return f
}

func formatFloat(f float64) string {
	return fmt.Sprintf("%.2f", f)
}
