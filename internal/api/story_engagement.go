package api

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"

	"privacy-social-backend/internal/repository/db"
)

type viewStoryRequest struct {
	StoryID string `uri:"id" binding:"required,uuid"`
}

// viewStory records that a user viewed a story
func (server *Server) viewStory(ctx *gin.Context) {
	var req viewStoryRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	authPayload := getAuthPayload(ctx)
	storyID, ok := parseUUIDParam(ctx, req.StoryID, "story_id")
	if !ok {
		return
	}

	// Check if user is the owner of the story. If so, do not record a view.
	// We need to fetch the story first to check ownership.
	story, err := server.store.GetStoryByID(ctx, storyID)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "story not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	// Check if blocked
	isBlocked, err := server.store.IsUserBlocked(ctx, db.IsUserBlockedParams{
		BlockerID: story.UserID,
		BlockedID: authPayload.UserID,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	if isBlocked {
		// Privacy: Act as if story doesn't exist or just forbid
		ctx.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	if story.UserID == authPayload.UserID {
		// Do not record view for own story
		ctx.JSON(http.StatusOK, gin.H{"message": "own story viewed"})
		return
	}

	view, err := server.store.CreateStoryView(ctx, db.CreateStoryViewParams{
		StoryID: storyID,
		UserID:  authPayload.UserID,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// Notify story owner via WebSocket
	event := struct {
		Type    string       `json:"type"`
		Payload db.StoryView `json:"payload"`
	}{
		Type:    "story_viewed",
		Payload: view,
	}

	eventBytes, err := json.Marshal(event)
	if err == nil {
		server.hub.SendToUser(story.UserID, eventBytes)
	} else {
		log.Error().Err(err).Msg("Failed to marshal story_viewed event")
	}

	ctx.JSON(http.StatusOK, view)
}

// getStoryViewers returns list of users who viewed the story (owner only)
func (server *Server) getStoryViewers(ctx *gin.Context) {
	var req viewStoryRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	authPayload := getAuthPayload(ctx)
	storyID, ok := parseUUIDParam(ctx, req.StoryID, "story_id")
	if !ok {
		return
	}

	// Verify user owns the story
	story, err := server.store.GetStoryByID(ctx, storyID)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "story not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	if story.UserID != authPayload.UserID {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "you can only view your own story viewers"})
		return
	}

	viewers, err := server.store.GetStoryViewers(ctx, storyID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// Viewers are now filtered by the SQL query to exclude the owner
	ctx.JSON(http.StatusOK, viewers)
}

type createReactionRequest struct {
	Emoji string `json:"emoji" binding:"required,min=1,max=10"`
}

// reactToStory adds or updates a reaction to a story
func (server *Server) reactToStory(ctx *gin.Context) {
	var uriReq viewStoryRequest
	if err := ctx.ShouldBindUri(&uriReq); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	var bodyReq createReactionRequest
	if err := ctx.ShouldBindJSON(&bodyReq); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	authPayload := getAuthPayload(ctx)
	storyID, ok := parseUUIDParam(ctx, uriReq.StoryID, "story_id")
	if !ok {
		return
	}

	reaction, err := server.store.CreateStoryReaction(ctx, db.CreateStoryReactionParams{
		StoryID: storyID,
		UserID:  authPayload.UserID,
		Emoji:   bodyReq.Emoji,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, reaction)
}

// deleteStoryReaction removes a user's reaction from a story
func (server *Server) deleteStoryReaction(ctx *gin.Context) {
	var req viewStoryRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	authPayload := getAuthPayload(ctx)
	storyID, ok := parseUUIDParam(ctx, req.StoryID, "story_id")
	if !ok {
		return
	}

	err := server.store.DeleteStoryReaction(ctx, db.DeleteStoryReactionParams{
		StoryID: storyID,
		UserID:  authPayload.UserID,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "reaction removed"})
}

// getStoryReactions returns all reactions for a story
func (server *Server) getStoryReactions(ctx *gin.Context) {
	var req viewStoryRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	storyID, ok := parseUUIDParam(ctx, req.StoryID, "story_id")
	if !ok {
		return
	}

	reactions, err := server.store.GetStoryReactions(ctx, storyID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, reactions)
}
