package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/rs/zerolog/log"

	"privacy-social-backend/internal/repository/db"
	"privacy-social-backend/internal/token"
)

type friendResponse struct {
	ID           uuid.UUID  `json:"id"`
	Username     string     `json:"username"`
	FullName     string     `json:"full_name"`
	AvatarUrl    string     `json:"avatar_url"`
	LastActiveAt *time.Time `json:"last_active_at"`
}

func (server *Server) listConnections(ctx *gin.Context) {
	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)

	connections, err := server.store.ListConnections(ctx, authPayload.UserID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	rsp := make([]friendResponse, len(connections))
	for i, c := range connections {
		var lastActive *time.Time
		if c.LastActiveAt.Valid {
			lastActive = &c.LastActiveAt.Time
		}
		rsp[i] = friendResponse{
			ID:           c.ID,
			Username:     c.Username,
			FullName:     c.FullName,
			AvatarUrl:    c.AvatarUrl.String,
			LastActiveAt: lastActive,
		}
	}

	ctx.JSON(http.StatusOK, rsp)
}

func (server *Server) listPendingRequests(ctx *gin.Context) {
	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)

	requests, err := server.store.ListPendingRequests(ctx, authPayload.UserID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, requests)
}

func (server *Server) listSentRequests(ctx *gin.Context) {
	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)

	requests, err := server.store.ListSentConnectionRequests(ctx, authPayload.UserID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, requests)
}

type connectionRequest struct {
	TargetUserID string `json:"target_user_id" binding:"required,uuid"`
}

func (server *Server) sendConnectionRequest(ctx *gin.Context) {
	var req connectionRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	targetID, err := uuid.Parse(req.TargetUserID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)

	if targetID == authPayload.UserID {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "cannot connect with yourself"})
		return
	}

	// Spam prevention: limit to 20 connection requests per day
	count, err := server.store.CountConnectionRequestsToday(ctx, authPayload.UserID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	if count >= 20 {
		ctx.JSON(http.StatusTooManyRequests, gin.H{"error": "daily connection request limit reached (20/day)"})
		return
	}

	// Get requester info for notification
	requester, err := server.store.GetUserByID(ctx, authPayload.UserID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	conn, err := server.store.CreateConnectionRequest(ctx, db.CreateConnectionRequestParams{
		RequesterID: authPayload.UserID,
		TargetID:    targetID,
	})
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code.Name() {
			case "unique_violation":
				// Idempotent success: if already exists, treat as success
				ctx.JSON(http.StatusOK, gin.H{"message": "connection request already sent"})
				return
			case "foreign_key_violation":
				ctx.JSON(http.StatusNotFound, gin.H{"error": "target user not found"})
				return
			}
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// Create notification for target user
	_, err = server.store.CreateNotification(ctx, db.CreateNotificationParams{
		UserID:        targetID,
		Type:          "connection_request",
		Title:         "New Connection Request",
		Message:       fmt.Sprintf("%s wants to connect with you", requester.Username),
		RelatedUserID: uuid.NullUUID{UUID: authPayload.UserID, Valid: true},
	})
	if err != nil {
		log.Error().Err(err).Msg("failed to create connection request notification")
	}

	ctx.JSON(http.StatusCreated, conn)
}

type updateConnectionRequest struct {
	RequesterID string `json:"requester_id" binding:"required,uuid"`
	Status      string `json:"status" binding:"required,oneof=accepted blocked"`
}

func (server *Server) updateConnection(ctx *gin.Context) {
	var req updateConnectionRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	requesterID, ok := parseUUIDParam(ctx, req.RequesterID, "requester_id")
	if !ok {
		return
	}
	authPayload := getAuthPayload(ctx)

	conn, err := server.store.UpdateConnectionStatus(ctx, db.UpdateConnectionStatusParams{
		RequesterID: requesterID,
		TargetID:    authPayload.UserID, // I am the target accepting the request
		Status:      db.ConnectionStatus(req.Status),
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// Create notification if connection was accepted
	if req.Status == "accepted" {
		accepter, err := server.store.GetUserByID(ctx, authPayload.UserID)
		if err == nil {
			_, err = server.store.CreateNotification(ctx, db.CreateNotificationParams{
				UserID:        requesterID,
				Type:          "connection_accepted",
				Title:         "Connection Accepted",
				Message:       fmt.Sprintf("%s accepted your connection request", accepter.Username),
				RelatedUserID: uuid.NullUUID{UUID: authPayload.UserID, Valid: true},
			})
			if err != nil {
				log.Error().Err(err).Msg("failed to create connection accepted notification")
			}
		}
	}

	ctx.JSON(http.StatusOK, conn)
}

func (server *Server) deleteConnection(ctx *gin.Context) {
	targetUserIDStr := ctx.Param("id")
	targetUserID, err := uuid.Parse(targetUserIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)

	err = server.store.DeleteConnection(ctx, db.DeleteConnectionParams{
		RequesterID: authPayload.UserID,
		TargetID:    targetUserID,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "connection deleted"})
}

type suggestedConnectionResponse struct {
	ID          uuid.UUID `json:"id"`
	Username    string    `json:"username"`
	FullName    string    `json:"full_name"`
	AvatarUrl   string    `json:"avatar_url"`
	MutualCount int64     `json:"mutual_count"`
}

func (server *Server) getSuggestedConnections(ctx *gin.Context) {
	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)

	suggestions, err := server.store.GetSuggestedConnections(ctx, db.GetSuggestedConnectionsParams{
		RequesterID: authPayload.UserID,
		Limit:       10,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	rsp := make([]suggestedConnectionResponse, len(suggestions))
	for i, s := range suggestions {
		rsp[i] = suggestedConnectionResponse{
			ID:          s.ID,
			Username:    s.Username,
			FullName:    s.FullName,
			AvatarUrl:   s.AvatarUrl.String,
			MutualCount: s.MutualCount,
		}
	}

	ctx.JSON(http.StatusOK, rsp)
}
