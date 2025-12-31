package api

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"privacy-social-backend/internal/repository/db"
)

type createReportRequest struct {
	TargetUserID  string `json:"target_user_id"`  // Optional
	TargetStoryID string `json:"target_story_id"` // Optional
	Reason        string `json:"reason" binding:"required,oneof=spam abuse inappropriate other"`
	Description   string `json:"description"`
}

func (server *Server) createReport(ctx *gin.Context) {
	var req createReportRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	authPayload := getAuthPayload(ctx)

	var targetUserID uuid.NullUUID
	if req.TargetUserID != "" {
		id, err := uuid.Parse(req.TargetUserID)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid target_user_id"})
			return
		}
		targetUserID = uuid.NullUUID{UUID: id, Valid: true}
	}

	var targetStoryID uuid.NullUUID
	if req.TargetStoryID != "" {
		id, err := uuid.Parse(req.TargetStoryID)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid target_story_id"})
			return
		}
		targetStoryID = uuid.NullUUID{UUID: id, Valid: true}
	}

	// Validate that at least one target is present
	if !targetUserID.Valid && !targetStoryID.Valid {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "must target user or story"})
		return
	}

	report, err := server.store.CreateReport(ctx, db.CreateReportParams{
		ReporterID:    authPayload.UserID,
		TargetUserID:  targetUserID,
		TargetStoryID: targetStoryID,
		Reason:        db.ReportReason(req.Reason),
		Description:   sql.NullString{String: req.Description, Valid: req.Description != ""},
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusCreated, report)
}
