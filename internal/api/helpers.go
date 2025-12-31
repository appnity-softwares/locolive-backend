package api

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"privacy-social-backend/internal/token"
)

// getAuthPayload extracts the authenticated user's payload from the context
func getAuthPayload(ctx *gin.Context) *token.Payload {
	return ctx.MustGet(authorizationPayloadKey).(*token.Payload)
}

// parseUUIDParam parses a UUID string and returns an error response if invalid
// Returns the parsed UUID and true if successful, or uuid.Nil and false if parsing failed
func parseUUIDParam(ctx *gin.Context, value string, paramName string) (uuid.UUID, bool) {
	id, err := uuid.Parse(value)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid %s", paramName)})
		return uuid.Nil, false
	}
	return id, true
}

// toNullString converts a string to a sql.NullString
func toNullString(s string) sql.NullString {
	return sql.NullString{
		String: s,
		Valid:  s != "",
	}
}

// nullStringToStrPtr converts a sql.NullString to a *string
func nullStringToStrPtr(ns sql.NullString) *string {
	if ns.Valid {
		return &ns.String
	}
	return nil
}
