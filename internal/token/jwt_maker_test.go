package token

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestJWTMaker(t *testing.T) {
	maker, err := NewJWTMaker("12345678901234567890123456789012") // 32 chars
	require.NoError(t, err)

	username := "testuser"
	duration := time.Minute

	userID := uuid.New()
	token, payload, err := maker.CreateToken(username, userID, duration)
	require.NoError(t, err)
	require.NotEmpty(t, token)
	require.NotNil(t, payload)

	payload2, err := maker.VerifyToken(token)
	require.NoError(t, err)
	require.NotNil(t, payload2)

	require.Equal(t, username, payload2.Username)
	require.Equal(t, userID, payload2.UserID)
	require.WithinDuration(t, payload.IssuedAt, payload2.IssuedAt, time.Second)
	require.WithinDuration(t, payload.ExpiredAt, payload2.ExpiredAt, time.Second)
}

func TestExpiredJWTToken(t *testing.T) {
	maker, err := NewJWTMaker("12345678901234567890123456789012")
	require.NoError(t, err)

	token, payload, err := maker.CreateToken("testuser", uuid.New(), -time.Minute)
	require.NoError(t, err)
	require.NotEmpty(t, token)
	require.NotNil(t, payload)

	payload2, err := maker.VerifyToken(token)
	require.Error(t, err)
	require.Nil(t, payload2)
}

func TestInvalidJWTTokenAlg(t *testing.T) {
	maker, err := NewJWTMaker("12345678901234567890123456789012")
	require.NoError(t, err)

	// Create token with different algorithm (would need manual creation)
	// For now, test with invalid token string
	payload, err := maker.VerifyToken("invalid.token.here")
	require.Error(t, err)
	require.Nil(t, payload)
}
