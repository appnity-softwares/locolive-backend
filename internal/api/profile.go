package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/sqlc-dev/pqtype"

	"privacy-social-backend/internal/repository/db"
	"privacy-social-backend/internal/token"
)

const profileCacheTTL = 10 * time.Minute

type UserLink struct {
	Label string `json:"label"`
	URL   string `json:"url"`
}

// ProfileResponse structure matching updated schema
type ProfileResponse struct {
	ID                uuid.UUID  `json:"id"`
	Username          string     `json:"username"`
	FullName          string     `json:"full_name"`
	AvatarUrl         string     `json:"avatar_url"`
	Bio               string     `json:"bio"`
	BannerUrl         string     `json:"banner_url"`
	Theme             string     `json:"theme"`
	ProfileVisibility string     `json:"profile_visibility"`
	Email             string     `json:"email"`
	IsGhostMode       bool       `json:"is_ghost_mode"`
	IsPremium         bool       `json:"is_premium"`
	ActivityStreak    int        `json:"activity_streak"`
	StoryCount        int64      `json:"story_count"`
	ConnectionCount   int64      `json:"connection_count"`
	LastActiveAt      time.Time  `json:"last_active_at"`
	CreatedAt         time.Time  `json:"created_at"`
	VisibilityStatus  string     `json:"visibility_status"`
	WebsiteURL        string     `json:"website_url"`
	Links             []UserLink `json:"links"`
}

func mapProfileResponse(p db.GetUserProfileRow) ProfileResponse {
	streak, _ := p.ActivityStreak.(int64)

	var links []UserLink
	if len(p.Links) > 0 {
		json.Unmarshal(p.Links, &links)
	}

	return ProfileResponse{
		ID:                p.ID,
		Username:          p.Username,
		FullName:          p.FullName,
		AvatarUrl:         p.AvatarUrl.String,
		Bio:               p.Bio.String,
		BannerUrl:         p.BannerUrl.String,
		Theme:             p.Theme.String,
		ProfileVisibility: p.ProfileVisibility.String,
		Email:             p.Email.String,
		IsGhostMode:       p.IsGhostMode,
		IsPremium:         p.IsPremium.Bool,
		ActivityStreak:    int(streak),
		StoryCount:        p.StoryCount,
		ConnectionCount:   p.ConnectionCount,
		LastActiveAt:      p.LastActiveAt.Time,
		CreatedAt:         p.CreatedAt,
		VisibilityStatus:  p.VisibilityStatus,
		WebsiteURL:        p.WebsiteUrl.String,
		Links:             links,
	}
}

// getUserProfile returns public profile information for any user by ID
func (server *Server) getUserProfile(ctx *gin.Context) {
	userIdStr := ctx.Param("id")
	userID, err := uuid.Parse(userIdStr)
	if err != nil {
		// Try resolving by username if UUID parse fails
		user, err := server.store.GetUserByUsername(ctx, userIdStr)
		if err != nil {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		userID = user.ID
	}

	// Track profile view if user is authenticated
	authPayload, exists := ctx.Get(authorizationPayloadKey)
	if exists && authPayload != nil {
		payload := authPayload.(*token.Payload)
		// Don't track self-views
		if payload.UserID != userID {
			// Track asynchronously to not block response
			go func() {
				server.store.TrackProfileView(context.Background(), db.TrackProfileViewParams{
					ViewerID:     payload.UserID,
					ViewedUserID: userID,
				})
			}()
		}
	}

	// Try Redis cache first
	cacheKey := "profile:" + userID.String()
	cachedData, err := server.redis.Get(context.Background(), cacheKey).Result()
	if err == nil && cachedData != "" {
		ctx.Header("X-Cache", "HIT")
		ctx.Data(http.StatusOK, "application/json", []byte(cachedData))
		return
	}

	profile, err := server.store.GetUserProfile(ctx, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	rsp := mapProfileResponse(profile)

	// Cache the result
	responseJSON, _ := json.Marshal(rsp)
	server.redis.Set(context.Background(), cacheKey, responseJSON, profileCacheTTL)

	ctx.Header("X-Cache", "MISS")
	ctx.JSON(http.StatusOK, rsp)
}

// getMyProfile returns the authenticated user's own profile
func (server *Server) getMyProfile(ctx *gin.Context) {
	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)

	// Try Redis cache first
	cacheKey := "profile:" + authPayload.UserID.String()
	cachedData, err := server.redis.Get(context.Background(), cacheKey).Result()
	if err == nil && cachedData != "" {
		ctx.Header("X-Cache", "HIT")
		ctx.Data(http.StatusOK, "application/json", []byte(cachedData))
		return
	}

	profile, err := server.store.GetUserProfile(ctx, authPayload.UserID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	rsp := mapProfileResponse(profile)

	// Cache the result
	responseJSON, _ := json.Marshal(rsp)
	server.redis.Set(context.Background(), cacheKey, responseJSON, profileCacheTTL)

	ctx.Header("X-Cache", "MISS")
	ctx.JSON(http.StatusOK, rsp)
}

type updateUserProfileRequest struct {
	FullName          string     `json:"full_name"`
	Username          string     `json:"username"`
	Bio               string     `json:"bio"`
	AvatarUrl         string     `json:"avatar_url"`
	BannerUrl         string     `json:"banner_url"`
	Theme             string     `json:"theme"`
	ProfileVisibility string     `json:"profile_visibility"`
	WebsiteURL        string     `json:"website_url"`
	Links             []UserLink `json:"links"`
}

func (server *Server) updateProfile(ctx *gin.Context) {
	var req updateUserProfileRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	payload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)
	arg := db.UpdateUserProfileParams{
		ID:                payload.UserID,
		FullName:          sql.NullString{String: req.FullName, Valid: req.FullName != ""},
		Username:          sql.NullString{String: req.Username, Valid: req.Username != ""},
		Bio:               sql.NullString{String: req.Bio, Valid: true},
		AvatarUrl:         sql.NullString{String: req.AvatarUrl, Valid: req.AvatarUrl != ""},
		BannerUrl:         sql.NullString{String: req.BannerUrl, Valid: req.BannerUrl != ""},
		Theme:             sql.NullString{String: req.Theme, Valid: req.Theme != ""},
		ProfileVisibility: sql.NullString{String: req.ProfileVisibility, Valid: req.ProfileVisibility != ""},
		WebsiteUrl:        sql.NullString{String: req.WebsiteURL, Valid: true},
	}

	if req.Links != nil {
		linksJSON, _ := json.Marshal(req.Links)
		arg.Links = pqtype.NullRawMessage{RawMessage: linksJSON, Valid: true}
	}

	user, err := server.store.UpdateUserProfile(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// Invalidate cache
	cacheKey := "profile:" + payload.UserID.String()
	server.redis.Del(context.Background(), cacheKey)

	// Return updated profile structure (simplified for update response)
	rsp := struct {
		ID                uuid.UUID  `json:"id"`
		Username          string     `json:"username"`
		FullName          string     `json:"full_name"`
		Bio               string     `json:"bio"`
		AvatarUrl         string     `json:"avatar_url"`
		BannerUrl         string     `json:"banner_url"`
		Theme             string     `json:"theme"`
		ProfileVisibility string     `json:"profile_visibility"`
		WebsiteUrl        string     `json:"website_url"`
		Links             []UserLink `json:"links"`
		CreatedAt         time.Time  `json:"created_at"`
	}{
		ID:                user.ID,
		Username:          user.Username,
		FullName:          user.FullName,
		Bio:               user.Bio.String,
		AvatarUrl:         user.AvatarUrl.String,
		BannerUrl:         user.BannerUrl.String,
		Theme:             user.Theme.String,
		ProfileVisibility: user.ProfileVisibility.String,
		WebsiteUrl:        user.WebsiteUrl.String,
		CreatedAt:         user.CreatedAt,
	}

	if len(user.Links) > 0 {
		json.Unmarshal(user.Links, &rsp.Links)
	}

	ctx.JSON(http.StatusOK, rsp)
}

// ProfileVisitorResponse represents a user who viewed the profile
type ProfileVisitorResponse struct {
	ID        uuid.UUID `json:"id"`
	Username  string    `json:"username"`
	FullName  string    `json:"full_name"`
	AvatarUrl string    `json:"avatar_url"`
	ViewedAt  time.Time `json:"viewed_at"`
}

// getProfileVisitors returns users who viewed the authenticated user's profile in last 24h
func (server *Server) getProfileVisitors(ctx *gin.Context) {
	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)

	visitors, err := server.store.GetRecentProfileVisitors(ctx, authPayload.UserID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	response := make([]ProfileVisitorResponse, len(visitors))
	for i, v := range visitors {
		response[i] = ProfileVisitorResponse{
			ID:        v.ID,
			Username:  v.Username,
			FullName:  v.FullName,
			AvatarUrl: v.AvatarUrl.String,
			ViewedAt:  v.ViewedAt,
		}
	}

	ctx.JSON(http.StatusOK, response)
}
