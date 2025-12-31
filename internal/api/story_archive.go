package api

import (
	"database/sql"
	"net/http"
	"strconv"

	"privacy-social-backend/internal/repository/db"
	"privacy-social-backend/internal/token"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// archiveStory saves a story to the user's archive before it expires
func (server *Server) archiveStory(ctx *gin.Context) {
	storyID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)

	// Archive the story
	archivedStory, err := server.store.ArchiveStory(ctx, db.ArchiveStoryParams{
		ID:     storyID,
		UserID: authPayload.UserID,
	})
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "story not found or already archived"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"message": "story archived successfully",
		"archive": archivedStory,
	})
}

// getArchivedStories returns all archived stories for the authenticated user
func (server *Server) getArchivedStories(ctx *gin.Context) {
	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)

	// Parse pagination parameters
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize

	// Get archived stories
	archives, err := server.store.GetArchivedStories(ctx, db.GetArchivedStoriesParams{
		UserID: authPayload.UserID,
		Limit:  int32(pageSize),
		Offset: int32(offset),
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// Get total count
	count, err := server.store.CountArchivedStories(ctx, authPayload.UserID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"archives":    archives,
		"total":       count,
		"page":        page,
		"page_size":   pageSize,
		"total_pages": (count + int64(pageSize) - 1) / int64(pageSize),
	})
}

// deleteArchivedStory removes a story from the user's archive
func (server *Server) deleteArchivedStory(ctx *gin.Context) {
	archiveID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)

	// Delete the archived story
	err = server.store.DeleteArchivedStory(ctx, db.DeleteArchivedStoryParams{
		ID:     archiveID,
		UserID: authPayload.UserID,
	})
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "archived story not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "archived story deleted successfully"})
}
