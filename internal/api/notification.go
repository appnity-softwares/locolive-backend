package api

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"

	"privacy-social-backend/internal/repository/db"
	"privacy-social-backend/internal/token"
)

type listNotificationsRequest struct {
	Page     int32 `form:"page" binding:"min=1"`
	PageSize int32 `form:"page_size" binding:"min=5,max=50"`
}

// getNotifications retrieves notifications for the authenticated user
func (server *Server) getNotifications(ctx *gin.Context) {
	var req listNotificationsRequest
	req.Page = 1
	req.PageSize = 20

	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)

	offset := (req.Page - 1) * req.PageSize
	notifications, err := server.store.ListNotifications(ctx, db.ListNotificationsParams{
		UserID: authPayload.UserID,
		Limit:  req.PageSize,
		Offset: offset,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, notifications)
}

type markNotificationReadRequest struct {
	NotificationID string `uri:"id" binding:"required,uuid"`
}

// markNotificationRead marks a notification as read
func (server *Server) markNotificationRead(ctx *gin.Context) {
	var req markNotificationReadRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	authPayload := getAuthPayload(ctx)

	notificationID, ok := parseUUIDParam(ctx, req.NotificationID, "notification_id")
	if !ok {
		return
	}

	notification, err := server.store.MarkNotificationAsRead(ctx, db.MarkNotificationAsReadParams{
		ID:     notificationID,
		UserID: authPayload.UserID,
	})
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "notification not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, notification)
}

// markAllNotificationsRead marks all notifications as read for the user
func (server *Server) markAllNotificationsRead(ctx *gin.Context) {
	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)

	err := server.store.MarkAllNotificationsAsRead(ctx, authPayload.UserID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "all notifications marked as read"})
}

// getUnreadCount returns the count of unread notifications
func (server *Server) getUnreadCount(ctx *gin.Context) {
	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)

	count, err := server.store.CountUnreadNotifications(ctx, authPayload.UserID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"unread_count": count})
}
