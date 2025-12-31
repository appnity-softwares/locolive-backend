package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"privacy-social-backend/internal/repository/db"
	"privacy-social-backend/internal/token"
)

// getActivityStatus returns the user's activity status and visibility
func (server *Server) getActivityStatus(ctx *gin.Context) {
	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)

	// Check if requesting for another user (Public/Story-based status)
	targetIDStr := ctx.Query("user_id")
	if targetIDStr != "" {
		targetID, err := uuid.Parse(targetIDStr)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, errorResponse(err))
			return
		}

		if targetID != authPayload.UserID {
			// 1. Get Base Status (for LastActiveAt)
			// We fetch this first to have the data. Privacy decides if we show it.
			targetStatus, err := server.store.GetUserActivityStatus(ctx, targetID)
			if err != nil {
				ctx.JSON(http.StatusInternalServerError, errorResponse(err))
				return
			}

			// 2. Check Public Logic (Story)
			hasStory, err := server.store.HasValidStory(ctx, targetID)
			if err != nil {
				ctx.JSON(http.StatusInternalServerError, errorResponse(err))
				return
			}

			// 3. Check Connection Logic (Friendship)
			// We allow seeing "Last Active" if you are connected
			conn, err := server.store.GetConnection(ctx, db.GetConnectionParams{
				RequesterID: authPayload.UserID,
				TargetID:    targetID,
			})
			isConnectedAndAccepted := err == nil && conn.Status == "accepted"

			// Determine Visibility
			isVisible := false
			visibilityStatus := "hidden"
			var lastActiveAt interface{} = nil

			// Logic:
			// If Has Story -> Active (Publicly visible ring)
			// If Connected -> Can see Last Active (even if not "active" right now)

			if hasStory {
				isVisible = true
				visibilityStatus = "active"
			}

			// 4. Check Real-time Online Status (Hub)
			isOnline := server.hub.IsUserOnline(targetID)

			if isConnectedAndAccepted {
				// If connected, we allow seeing the timestamp
				if targetStatus.LastActiveAt.Valid {
					lastActiveAt = targetStatus.LastActiveAt.Time
				} else {
					lastActiveAt = nil
				}
			}

			if isVisible || lastActiveAt != nil || isOnline {
				ctx.JSON(http.StatusOK, gin.H{
					"visibility_status": visibilityStatus, // controls Story Ring
					"is_visible":        true,
					"id":                targetID,
					"last_active_at":    lastActiveAt,
					"is_online":         isOnline, // Controls "Active Now" text/dot
				})
			} else {
				ctx.JSON(http.StatusOK, gin.H{
					"visibility_status": "hidden",
					"is_visible":        false,
					"id":                targetID,
				})
			}
			return
		}
	}

	// Default: Get Own Status (Full details)
	status, err := server.store.GetUserActivityStatus(ctx, authPayload.UserID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, status)
}
