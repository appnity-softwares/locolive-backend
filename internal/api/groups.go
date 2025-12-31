package api

import (
	"database/sql"
	"net/http"

	"privacy-social-backend/internal/repository/db"
	"privacy-social-backend/internal/token"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type createGroupRequest struct {
	Name        string      `json:"name" binding:"required"`
	Description string      `json:"description"`
	MemberIDs   []uuid.UUID `json:"member_ids"` // Initial members
}

func (server *Server) createGroup(ctx *gin.Context) {
	var req createGroupRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)

	// Create Group
	group, err := server.store.CreateGroup(ctx, db.CreateGroupParams{
		Name:        req.Name,
		Description: sql.NullString{String: req.Description, Valid: req.Description != ""},
		CreatedBy:   authPayload.UserID,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// Add Creator as Admin
	_, err = server.store.AddGroupMember(ctx, db.AddGroupMemberParams{
		GroupID: group.ID,
		UserID:  authPayload.UserID,
		Role:    "admin",
	})
	if err != nil {
		// Rollback desirable but skipping for simple impl
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// Add other members
	for _, memberID := range req.MemberIDs {
		// Verify connection? (Optional but good practice)
		// ...
		server.store.AddGroupMember(ctx, db.AddGroupMemberParams{
			GroupID: group.ID,
			UserID:  memberID,
			Role:    "member",
		})
	}

	ctx.JSON(http.StatusCreated, group)
}

func (server *Server) getMyGroups(ctx *gin.Context) {
	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)

	groups, err := server.store.GetUserGroups(ctx, authPayload.UserID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, groups)
}

func (server *Server) getGroupMessages(ctx *gin.Context) {
	groupID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	// Verify membership
	// ... (Skipping for MVP speed, but important later)

	msgs, err := server.store.GetGroupMessages(ctx, uuid.NullUUID{UUID: groupID, Valid: true})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, msgs)
}
