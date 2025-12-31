package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mmcloughlin/geohash"

	"privacy-social-backend/internal/repository/db"
	"privacy-social-backend/internal/service/story"
	"privacy-social-backend/internal/token"
)

const (
	defaultRadiusMeters = 5000  // 5km
	maxRadiusMeters     = 20000 // 20km
	radiusStepMeters    = 5000  // 5km step
	feedCacheTTL        = 5 * time.Minute
)

type createStoryRequest struct {
	MediaURL     string  `json:"media_url" binding:"required"`
	MediaType    string  `json:"media_type" binding:"required,oneof=image video text"`
	Latitude     float64 `json:"latitude" binding:"required,min=-90,max=90"`
	Longitude    float64 `json:"longitude" binding:"required,min=-180,max=180"`
	Caption      string  `json:"caption"`
	IsAnonymous  bool    `json:"is_anonymous"`
	ShowLocation bool    `json:"show_location"`
}

func (server *Server) createStory(ctx *gin.Context) {
	var req createStoryRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)

	result, err := server.story.CreateStory(ctx, story.CreateStoryParams{
		UserID:       authPayload.UserID,
		MediaURL:     req.MediaURL,
		MediaType:    req.MediaType,
		Latitude:     req.Latitude,
		Longitude:    req.Longitude,
		Caption:      req.Caption,
		IsAnonymous:  req.IsAnonymous,
		ShowLocation: req.ShowLocation,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusCreated, toStoryResponseFromCreate(*result))
}

type getFeedRequest struct {
	Latitude  float64 `form:"latitude" binding:"required,min=-90,max=90"`
	Longitude float64 `form:"longitude" binding:"required,min=-180,max=180"`
}

func (server *Server) getFeed(ctx *gin.Context) {
	var req getFeedRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)

	// Create cache key based on user's geohash (5 chars = ~2.4km precision)
	userGeohash := geohash.Encode(req.Latitude, req.Longitude)
	if len(userGeohash) > 5 {
		userGeohash = userGeohash[:5]
	}
	cacheKey := "feed:" + userGeohash

	// Try to get from Redis cache first
	cachedData, err := server.redis.Get(ctx, cacheKey).Result()
	if err == nil && cachedData != "" {
		// Cache hit - return cached data
		ctx.Header("X-Cache", "HIT")
		ctx.Data(http.StatusOK, "application/json", []byte(cachedData))
		return
	}

	stories, message, radius, err := server.story.GetFeed(ctx, story.GetFeedParams{
		UserID:    authPayload.UserID,
		Latitude:  req.Latitude,
		Longitude: req.Longitude,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// Convert to response DTOs
	storyResponses := make([]StoryResponse, len(stories))
	for i, story := range stories {
		storyResponses[i] = toStoryResponse(story)
	}

	response := gin.H{
		"stories":       storyResponses,
		"count":         len(storyResponses),
		"message":       message,
		"search_radius": radius,
	}

	// Cache the result for 5 minutes
	responseJSON, _ := json.Marshal(response)
	server.redis.Set(ctx, cacheKey, responseJSON, feedCacheTTL)

	ctx.Header("X-Cache", "MISS")
	ctx.JSON(http.StatusOK, response)
}

// deleteStory allows users to delete their own stories
func (server *Server) deleteUserStory(ctx *gin.Context) {
	storyID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)

	err = server.story.DeleteStory(ctx, storyID, authPayload.UserID)
	if err != nil {
		if err.Error() == "story not found" {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "story not found"})
			return
		}
		if err.Error() == "you can only delete your own stories" {
			ctx.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}

		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "story deleted successfully"})
}

type updateStoryRequest struct {
	Caption      *string `json:"caption"`
	IsAnonymous  *bool   `json:"is_anonymous"`
	ShowLocation *bool   `json:"show_location"`
}

// updateStory allows users to edit their story within 15 minutes of posting
func (server *Server) updateStory(ctx *gin.Context) {
	storyID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	var req updateStoryRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)

	// Prepare nullable parameters for SQL
	var captionArg sql.NullString
	if req.Caption != nil {
		captionArg = sql.NullString{String: *req.Caption, Valid: true}
	}

	var isAnonymousArg sql.NullBool
	if req.IsAnonymous != nil {
		isAnonymousArg = sql.NullBool{Bool: *req.IsAnonymous, Valid: true}
	}

	var showLocationArg sql.NullBool
	if req.ShowLocation != nil {
		showLocationArg = sql.NullBool{Bool: *req.ShowLocation, Valid: true}
	}

	// Update the story
	story, err := server.store.UpdateStory(ctx, db.UpdateStoryParams{
		ID:           storyID,
		UserID:       authPayload.UserID,
		Caption:      captionArg,
		IsAnonymous:  isAnonymousArg,
		ShowLocation: showLocationArg,
	})
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusForbidden, gin.H{"error": "story not found, expired, or edit window closed (15 minutes)"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// Invalidate feed cache
	userGeohash := story.Geohash
	if len(userGeohash) > 5 {
		userGeohash = userGeohash[:5]
	}
	server.invalidateFeedCache(userGeohash)

	// Convert to response
	rsp := toStoryResponseFromUpdate(story)

	ctx.JSON(http.StatusOK, rsp)
}

// getConnectionStories returns stories from connected users, ignoring radius
func (server *Server) getConnectionStories(ctx *gin.Context) {
	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)

	// Cache key based on user ID
	cacheKey := "stories:connections:" + authPayload.UserID.String()

	// Try Redis cache first
	cachedData, err := server.redis.Get(ctx, cacheKey).Result()
	if err == nil && cachedData != "" {
		ctx.Header("X-Cache", "HIT")
		ctx.Data(http.StatusOK, "application/json", []byte(cachedData))
		return
	}

	stories, err := server.store.GetConnectionStories(ctx, authPayload.UserID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// Convert to response DTOs
	storyResponses := make([]StoryResponse, len(stories))
	for i, story := range stories {
		storyResponses[i] = toStoryResponseFromConnection(story)
	}

	// Cache for 5 minutes
	responseJSON, _ := json.Marshal(storyResponses)
	server.redis.Set(ctx, cacheKey, responseJSON, feedCacheTTL)

	ctx.Header("X-Cache", "MISS")
	ctx.JSON(http.StatusOK, storyResponses)
}

// getStory retrieves a single story by ID
func (server *Server) getStory(ctx *gin.Context) {
	storyID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	story, err := server.store.GetStoryByID(ctx, storyID)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "story not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// Check if story is expired
	if time.Now().After(story.ExpiresAt) {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "story has expired"})
		return
	}

	// Convert to response DTO
	rsp := toStoryResponseFromGet(story)

	// Fetch author details since they aren't in the partial story object
	user, err := server.store.GetUserByID(ctx, story.UserID)
	if err == nil {
		rsp.Username = user.Username
		if user.AvatarUrl.Valid {
			rsp.AvatarURL = &user.AvatarUrl.String
		}
	}

	ctx.JSON(http.StatusOK, rsp)
}
