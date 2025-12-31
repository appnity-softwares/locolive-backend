package api

import (
	"database/sql"
	"net/http"
	"time"

	"privacy-social-backend/internal/repository/db"
	"privacy-social-backend/internal/token"

	"github.com/gin-gonic/gin"
)

// boostProfile activates a 24-hour discovery boost for the user
func (server *Server) boostProfile(ctx *gin.Context) {
	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)

	// Check if user is premium
	user, err := server.store.GetUserByID(ctx, authPayload.UserID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// Simple & Fair: Only Premium users can boost (Perk)
	// Or users can pay for it? User requirement said "Optional profile boost".
	// We'll enforce Premium requirement for now as "Simple" logic.
	if !user.IsPremium.Valid || !user.IsPremium.Bool {
		err := map[string]string{"error": "Profile boost is a Premium feature"}
		ctx.JSON(http.StatusForbidden, err)
		return
	}

	// Activate 24h Boost
	expiresAt := time.Now().UTC().Add(24 * time.Hour)

	_, err = server.store.BoostUser(ctx, db.BoostUserParams{
		ID:             authPayload.UserID,
		BoostExpiresAt: sql.NullTime{Time: expiresAt, Valid: true},
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status":     "boosted",
		"expires_at": expiresAt,
		"message":    "Your profile is now boosted for 24 hours!",
	})
}
