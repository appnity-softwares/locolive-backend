package api

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"privacy-social-backend/internal/repository/db"
	"privacy-social-backend/internal/util"

	"github.com/gin-gonic/gin"
)

type forgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

func (server *Server) forgotPassword(ctx *gin.Context) {
	var req forgotPasswordRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	_, err := server.store.GetUserByEmail(ctx, sql.NullString{String: req.Email, Valid: true})
	if err != nil {
		if err == sql.ErrNoRows {
			// Do not reveal email existence
			ctx.JSON(http.StatusOK, gin.H{"message": "If this email exists, a reset link has been sent."})
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// Generate Token
	resetToken := util.RandomString(32)
	expiresAt := time.Now().Add(15 * time.Minute)

	_, err = server.store.SetPasswordResetToken(ctx, db.SetPasswordResetTokenParams{
		Email:                  sql.NullString{String: req.Email, Valid: true},
		PasswordResetToken:     sql.NullString{String: resetToken, Valid: true},
		PasswordResetExpiresAt: sql.NullTime{Time: expiresAt, Valid: true},
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// LOG TOKEN (For development since no SMTP)
	fmt.Printf("------------\n[PASSWORD RESET]\nUser: %s\nToken: %s\n------------\n", req.Email, resetToken)

	ctx.JSON(http.StatusOK, gin.H{"message": "If this email exists, a reset link has been sent."})
}

type resetPasswordRequest struct {
	Token       string `json:"token" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

func (server *Server) resetPassword(ctx *gin.Context) {
	var req resetPasswordRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	user, err := server.store.GetUserByResetToken(ctx, sql.NullString{String: req.Token, Valid: true})
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
		return
	}

	hashedPassword, err := util.HashPassword(req.NewPassword)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// Update Password
	err = server.store.UpdateUserPassword(ctx, db.UpdateUserPasswordParams{
		ID:           user.ID,
		PasswordHash: hashedPassword,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// Clear Token
	server.store.ClearPasswordResetToken(ctx, user.ID)

	ctx.JSON(http.StatusOK, gin.H{"message": "password updated successfully"})
}
