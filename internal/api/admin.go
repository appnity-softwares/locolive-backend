package api

import (
	"net/http"
	"privacy-social-backend/internal/service/admin"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	adminStatsCacheTTL = 1 * time.Minute
)

// Admin: List Users
type listUsersRequest struct {
	PageID   int32 `form:"page" binding:"required,min=1"`
	PageSize int32 `form:"page_size" binding:"required,min=5,max=100"`
}

func (server *Server) listUsers(ctx *gin.Context) {
	var req listUsersRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	users, count, err := server.admin.ListUsers(ctx, admin.ListUsersParams{
		PageID:   req.PageID,
		PageSize: req.PageSize,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"users": users,
		"total": count,
		"page":  req.PageID,
	})
}

// Admin: Ban/Unban User
type banUserRequest struct {
	UserID string `json:"user_id" binding:"required,uuid"`
	Ban    bool   `json:"ban" binding:"required"`
}

func (server *Server) banUser(ctx *gin.Context) {
	var req banUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	_, ok := parseUUIDParam(ctx, req.UserID, "user_id")
	if !ok {
		return
	}

	// Wait, req.UserID is string in my param struct but the parser helper returns uuid.
	// Actually banUserRequest defined UserID as string.
	// Service expects BanUserParams with string UserID to be parsed inside or UUID?
	// AdminService.BanUser takes params.UserID as string and parses it.
	// So I can pass req.UserID directly if I want, or pass the UUID.
	// Let's check AdminService signature again.
	// func (s *ServiceImpl) BanUser(ctx context.Context, params BanUserParams) (db.User, error)
	// type BanUserParams struct { UserID string; Ban bool }

	// So I should pass string. parseUUIDParam is doing validation which is good.
	// I will keep validation but just pass the string to service.

	user, err := server.admin.BanUser(ctx, admin.BanUserParams{
		UserID: req.UserID,
		Ban:    req.Ban,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, user)
}

// Admin: Delete User
type deleteUserRequest struct {
	UserID string `uri:"id" binding:"required,uuid"`
}

func (server *Server) deleteUser(ctx *gin.Context) {
	var req deleteUserRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	err := server.admin.DeleteUser(ctx, req.UserID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "user deleted"})
}

// Admin: Get Statistics (with Redis caching)
func (server *Server) getStats(ctx *gin.Context) {
	response, isCached, err := server.admin.GetStats(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	if isCached {
		ctx.Header("X-Cache", "HIT")
	} else {
		ctx.Header("X-Cache", "MISS")
	}
	ctx.JSON(http.StatusOK, response)
}

// Admin: List Reports
type listReportsRequest struct {
	Resolved bool  `form:"resolved"`
	PageID   int32 `form:"page" binding:"required,min=1"`
	PageSize int32 `form:"page_size" binding:"required,min=5,max=100"`
}

func (server *Server) listReports(ctx *gin.Context) {
	var req listReportsRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	reports, err := server.admin.ListReports(ctx, req.Resolved, req.PageID, req.PageSize)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, reports)
}

// Admin: Resolve Report
type resolveReportRequest struct {
	ReportID string `uri:"id" binding:"required,uuid"`
}

func (server *Server) resolveReport(ctx *gin.Context) {
	var req resolveReportRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	report, err := server.admin.ResolveReport(ctx, req.ReportID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, report)
}

// Admin: Delete Story
type deleteStoryRequest struct {
	StoryID string `uri:"id" binding:"required,uuid"`
}

func (server *Server) deleteStory(ctx *gin.Context) {
	var req deleteStoryRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	err := server.admin.DeleteStory(ctx, req.StoryID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "story deleted"})
}

// Admin: List All Stories
type listAllStoriesRequest struct {
	PageID   int32 `form:"page" binding:"required,min=1"`
	PageSize int32 `form:"page_size" binding:"required,min=5,max=100"`
}

func (server *Server) listAllStories(ctx *gin.Context) {
	var req listAllStoriesRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	stories, err := server.admin.ListAllStories(ctx, req.PageID, req.PageSize)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, stories)
}
