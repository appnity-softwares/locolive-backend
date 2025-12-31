package token

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const minSecretKeySize = 32

// JWTMaker is a JSON Web Token maker
type JWTMaker struct {
	secretKey string
}

// NewJWTMaker creates a new JWTMaker
func NewJWTMaker(secretKey string) (Maker, error) {
	if len(secretKey) < minSecretKeySize {
		return nil, fmt.Errorf("invalid key size: must be at least %d characters", minSecretKeySize)
	}
	return &JWTMaker{secretKey}, nil
}

// CreateToken creates a new token for a specific username and duration
func (maker *JWTMaker) CreateToken(username string, userID uuid.UUID, duration time.Duration) (string, *Payload, error) {
	payload, err := NewPayload(username, userID, duration)
	if err != nil {
		return "", payload, err
	}

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":         payload.ID.String(),
		"user_id":    payload.UserID.String(),
		"username":   payload.Username,
		"issued_at":  payload.IssuedAt.Format(time.RFC3339Nano),
		"expired_at": payload.ExpiredAt.Format(time.RFC3339Nano),
	})

	token, err := jwtToken.SignedString([]byte(maker.secretKey))
	return token, payload, err
}

// VerifyToken checks if the token is valid or not
func (maker *JWTMaker) VerifyToken(token string) (*Payload, error) {
	keyFunc := func(token *jwt.Token) (interface{}, error) {
		_, ok := token.Method.(*jwt.SigningMethodHMAC)
		if !ok {
			return nil, ErrInvalidToken
		}
		return []byte(maker.secretKey), nil
	}

	jwtToken, err := jwt.Parse(token, keyFunc)
	if err != nil {
		return nil, err
	}

	claims, ok := jwtToken.Claims.(jwt.MapClaims)
	if !ok || !jwtToken.Valid {
		return nil, ErrInvalidToken
	}

	// Parse UUID from string
	idStr, ok := claims["id"].(string)
	if !ok {
		return nil, ErrInvalidToken
	}
	id, err := uuid.Parse(idStr)
	if err != nil {
		return nil, ErrInvalidToken
	}

	// Parse UserID
	userIDStr, ok := claims["user_id"].(string)
	if !ok {
		return nil, ErrInvalidToken
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, ErrInvalidToken
	}

	// Parse username
	username, ok := claims["username"].(string)
	if !ok {
		return nil, ErrInvalidToken
	}

	// Parse issued_at
	issuedAtStr, ok := claims["issued_at"].(string)
	if !ok {
		return nil, ErrInvalidToken
	}
	issuedAt, err := time.Parse(time.RFC3339Nano, issuedAtStr)
	if err != nil {
		return nil, ErrInvalidToken
	}

	// Parse expired_at
	expiredAtStr, ok := claims["expired_at"].(string)
	if !ok {
		return nil, ErrInvalidToken
	}
	expiredAt, err := time.Parse(time.RFC3339Nano, expiredAtStr)
	if err != nil {
		return nil, ErrInvalidToken
	}

	payload := &Payload{
		ID:        id,
		UserID:    userID,
		Username:  username,
		IssuedAt:  issuedAt,
		ExpiredAt: expiredAt,
	}

	// Check if token is expired
	if err := payload.Valid(); err != nil {
		return nil, err
	}

	return payload, nil
}
