package api

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"privacy-social-backend/internal/repository/db"
	"privacy-social-backend/internal/token"
)

// Privacy Settings Handlers

type PrivacySettingResponse struct {
	UserID           uuid.UUID `json:"user_id"`
	WhoCanMessage    string    `json:"who_can_message"`
	WhoCanSeeStories string    `json:"who_can_see_stories"`
	ShowLocation     bool      `json:"show_location"`
}

func newPrivacySettingResponse(p db.PrivacySetting) PrivacySettingResponse {
	return PrivacySettingResponse{
		UserID:           p.UserID,
		WhoCanMessage:    p.WhoCanMessage.String,
		WhoCanSeeStories: p.WhoCanSeeStories.String,
		ShowLocation:     p.ShowLocation.Bool,
	}
}

type updatePrivacySettingsRequest struct {
	WhoCanMessage    string `json:"who_can_message" binding:"oneof=everyone connections nobody"`
	WhoCanSeeStories string `json:"who_can_see_stories" binding:"oneof=everyone connections nobody"`
	ShowLocation     *bool  `json:"show_location" binding:"required"`
}

func (server *Server) updatePrivacySettings(ctx *gin.Context) {
	var req updatePrivacySettingsRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	payload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)

	settings, err := server.store.UpsertPrivacySettings(ctx, db.UpsertPrivacySettingsParams{
		UserID:           payload.UserID,
		WhoCanMessage:    sql.NullString{String: req.WhoCanMessage, Valid: true},
		WhoCanSeeStories: sql.NullString{String: req.WhoCanSeeStories, Valid: true},
		ShowLocation:     sql.NullBool{Bool: *req.ShowLocation, Valid: true},
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, newPrivacySettingResponse(settings))
}

func (server *Server) getPrivacySettings(ctx *gin.Context) {
	payload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)

	settings, err := server.store.GetPrivacySettings(ctx, payload.UserID)
	if err != nil {
		if err == sql.ErrNoRows {
			// Return default settings if none exist
			ctx.JSON(http.StatusOK, PrivacySettingResponse{
				UserID:           payload.UserID,
				WhoCanMessage:    "connections",
				WhoCanSeeStories: "connections",
				ShowLocation:     true,
			})
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, newPrivacySettingResponse(settings))
}

// Blocking Handlers

type blockUserRequest struct {
	UserID string `json:"user_id" binding:"required,uuid"`
}

func (server *Server) blockUser(ctx *gin.Context) {
	var req blockUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	payload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)
	blockID, ok := parseUUIDParam(ctx, req.UserID, "user_id")
	if !ok {
		return
	}

	// Prevent blocking self
	if payload.UserID == blockID {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "cannot block yourself"})
		return
	}

	_, err := server.store.BlockUser(ctx, db.BlockUserParams{
		BlockerID: payload.UserID,
		BlockedID: blockID,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// Invalidate caches
	server.invalidateProfileCache(payload.UserID)
	server.invalidateProfileCache(blockID)
	server.redis.Del(context.Background(), "connections:"+payload.UserID.String())

	ctx.JSON(http.StatusOK, gin.H{"message": "user blocked"})
}

func (server *Server) unblockUser(ctx *gin.Context) {
	targetIDStr := ctx.Param("id")
	targetID, ok := parseUUIDParam(ctx, targetIDStr, "user_id")
	if !ok {
		return
	}

	payload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)

	err := server.store.UnblockUser(ctx, db.UnblockUserParams{
		BlockerID: payload.UserID,
		BlockedID: targetID,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "user unblocked"})
}

type BlockedUserResponse struct {
	ID        uuid.UUID `json:"id"`
	Username  string    `json:"username"`
	FullName  string    `json:"full_name"`
	AvatarUrl string    `json:"avatar_url"`
	BlockedAt time.Time `json:"blocked_at"`
}

func (server *Server) getBlockedUsers(ctx *gin.Context) {
	payload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)

	users, err := server.store.GetBlockedUsers(ctx, payload.UserID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	rsp := make([]BlockedUserResponse, len(users))
	for i, u := range users {
		rsp[i] = BlockedUserResponse{
			ID:        u.ID,
			Username:  u.Username,
			FullName:  u.FullName,
			AvatarUrl: u.AvatarUrl.String,
			BlockedAt: u.BlockedAt.Time,
		}
	}

	ctx.JSON(http.StatusOK, rsp)
}

// Location Privacy

type toggleGhostModeRequest struct {
	Enabled  bool `json:"enabled"`
	Duration int  `json:"duration"` // minutes, 0 = indefinite
}

func (server *Server) toggleGhostMode(ctx *gin.Context) {
	var req toggleGhostModeRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	payload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)

	var expiresAt sql.NullTime
	if req.Enabled && req.Duration > 0 {
		expiresAt = sql.NullTime{
			Time:  time.Now().Add(time.Duration(req.Duration) * time.Minute),
			Valid: true,
		}
	}

	// Call existing ToggleGhostMode query - it returns the updated user
	user, err := server.store.ToggleGhostMode(ctx, db.ToggleGhostModeParams{
		ID:                 payload.UserID,
		IsGhostMode:        req.Enabled,
		GhostModeExpiresAt: expiresAt,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// Return the updated user object so frontend gets fresh data
	ctx.JSON(http.StatusOK, newUserResponse(user))
}

func (server *Server) panicMode(ctx *gin.Context) {
	payload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)

	// Delete all user data
	err := server.store.DeleteAllUserData(ctx, payload.UserID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// Invalidate token/session would be good here but handled by expiry usually

	ctx.JSON(http.StatusOK, gin.H{"message": "all data deleted"})
}
