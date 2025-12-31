package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mmcloughlin/geohash"

	"privacy-social-backend/internal/repository/db"
	"privacy-social-backend/internal/token"
)

type getStoriesMapRequest struct {
	North float64 `form:"north" binding:"required,min=-90,max=90"`
	South float64 `form:"south" binding:"required,min=-90,max=90"`
	East  float64 `form:"east" binding:"required,min=-180,max=180"`
	West  float64 `form:"west" binding:"required,min=-180,max=180"`
}

const mapCacheTTL = 5 * time.Minute

// getStoriesMap returns stories within a bounding box for map display
func (server *Server) getStoriesMap(ctx *gin.Context) {
	var req getStoriesMapRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	// Get auth payload for privacy/block rules
	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)

	// Validate bounding box
	if req.North <= req.South {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "north must be greater than south"})
		return
	}
	if req.East <= req.West {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "east must be greater than west"})
		return
	}

	// Create cache key from bounding box (rounded to 2 decimals for better cache hits) + UserID for personalization
	cacheKey := fmt.Sprintf("map:%.2f:%.2f:%.2f:%.2f:%s", req.North, req.South, req.East, req.West, authPayload.UserID)

	// Try Redis cache first
	cachedData, err := server.redis.Get(context.Background(), cacheKey).Result()
	if err == nil && cachedData != "" {
		ctx.Header("X-Cache", "HIT")
		ctx.Data(http.StatusOK, "application/json", []byte(cachedData))
		return
	}

	stories, err := server.store.GetStoriesInBounds(ctx, db.GetStoriesInBoundsParams{
		North:         req.North,
		South:         req.South,
		East:          req.East,
		West:          req.West,
		CurrentUserID: authPayload.UserID,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// Cluster stories by geohash (5 chars = ~2.4km precision)
	clusters := make(map[string][]db.GetStoriesInBoundsRow)
	for _, story := range stories {
		// Get 5-char geohash for clustering
		hash := story.Geohash
		if len(hash) > 5 {
			hash = hash[:5]
		}
		clusters[hash] = append(clusters[hash], story)
	}

	// Convert clusters to response format
	type ClusterResponse struct {
		Geohash   string          `json:"geohash"`
		Latitude  float64         `json:"latitude"`
		Longitude float64         `json:"longitude"`
		Count     int             `json:"count"`
		Stories   []StoryResponse `json:"stories,omitempty"`
	}

	var response []ClusterResponse
	for hash, clusterStories := range clusters {
		lat, lng := geohash.Decode(hash)

		cluster := ClusterResponse{
			Geohash:   hash,
			Latitude:  lat,
			Longitude: lng,
			Count:     len(clusterStories),
		}

		// If cluster has 3 or fewer stories, include them
		// Otherwise just show count for privacy
		if len(clusterStories) <= 3 {
			cluster.Stories = make([]StoryResponse, len(clusterStories))
			for i, story := range clusterStories {
				cluster.Stories[i] = toStoryResponseFromBounds(story)
			}
		}

		response = append(response, cluster)
	}

	result := gin.H{
		"clusters": response,
		"total":    len(stories),
	}

	// Cache the result
	responseJSON, _ := json.Marshal(result)
	server.redis.Set(context.Background(), cacheKey, responseJSON, mapCacheTTL)

	ctx.Header("X-Cache", "MISS")
	ctx.JSON(http.StatusOK, result)
}
