package api

import (
	"context"
	"encoding/json"
	"net/http"
	"sort"
	"time"

	"github.com/gin-gonic/gin"

	"privacy-social-backend/internal/token"
)

const crossingsCacheTTL = 5 * time.Minute

// CrossingResponse defines the API response for a path crossing
type CrossingResponse struct {
	ID             string    `json:"id"`
	UserID         string    `json:"user_id"`
	Username       string    `json:"username"`
	FullName       string    `json:"full_name"`
	AvatarURL      string    `json:"avatar_url"`
	LastCrossingAt time.Time `json:"last_crossing_at"`
	CrossingCount  int       `json:"crossing_count"`
}

// getCrossings returns crossings for the authenticated user
func (server *Server) getCrossings(ctx *gin.Context) {
	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)

	// Try Redis cache first
	cacheKey := "crossings:v3:" + authPayload.UserID.String()
	cachedData, err := server.redis.Get(context.Background(), cacheKey).Result()
	if err == nil && cachedData != "" {
		ctx.Header("X-Cache", "HIT")
		ctx.Data(http.StatusOK, "application/json", []byte(cachedData))
		return
	}

	crossings, err := server.store.GetCrossingsForUser(ctx, authPayload.UserID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// Aggregate by other user ID
	grouped := make(map[string]*CrossingResponse)

	for _, c := range crossings {
		var otherUserID string
		if c.UserID1 == authPayload.UserID {
			otherUserID = c.UserID2.String()
		} else {
			otherUserID = c.UserID1.String()
		}

		if existing, found := grouped[otherUserID]; found {
			// Update existing
			existing.CrossingCount++
			if c.OccurredAt.After(existing.LastCrossingAt) {
				existing.LastCrossingAt = c.OccurredAt
			}
		} else {
			// New entry - fetch user details
			// Fetch other user details
			user, err := server.store.GetUserByID(ctx, c.UserID1)
			if c.UserID1 == authPayload.UserID {
				user, err = server.store.GetUserByID(ctx, c.UserID2)
			}

			if err != nil {
				continue // Skip if user not found
			}

			grouped[otherUserID] = &CrossingResponse{
				ID:             c.ID.String(),
				UserID:         otherUserID,
				Username:       user.Username,
				FullName:       user.FullName,
				AvatarURL:      user.AvatarUrl.String,
				LastCrossingAt: c.OccurredAt,
				CrossingCount:  1,
			}
		}
	}

	// Convert map to slice
	var response []CrossingResponse
	for _, v := range grouped {
		response = append(response, *v)
	}

	// Sort by Recent first
	sort.Slice(response, func(i, j int) bool {
		return response[i].LastCrossingAt.After(response[j].LastCrossingAt)
	})

	// Cache the result
	marshaled, err := json.Marshal(response)
	if err == nil {
		server.redis.Set(context.Background(), cacheKey, marshaled, crossingsCacheTTL)
	}

	ctx.JSON(http.StatusOK, response)
}
