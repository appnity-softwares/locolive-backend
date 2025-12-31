package user

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"

	"privacy-social-backend/internal/repository"
	"privacy-social-backend/internal/repository/db"
	"privacy-social-backend/internal/token"
	"privacy-social-backend/internal/util"
)

type CreateUserParams struct {
	Phone    string
	Username string
	FullName string
	Password string
}

type LoginUserParams struct {
	Phone     string
	Password  string
	UserAgent string
	ClientIP  string
}

type LoginUserResult struct {
	SessionID             uuid.UUID
	AccessToken           string
	AccessTokenExpiresAt  time.Time
	RefreshToken          string
	RefreshTokenExpiresAt time.Time
	User                  db.User
}

type UpdateEmailParams struct {
	UserID uuid.UUID
	Email  string
}

type Service interface {
	CreateUser(ctx context.Context, params CreateUserParams) (db.User, error)
	LoginUser(ctx context.Context, params LoginUserParams) (*LoginUserResult, error)
	UpdateEmail(ctx context.Context, params UpdateEmailParams) (db.User, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (db.User, error)
	UpdatePassword(ctx context.Context, userID uuid.UUID, currentPassword, newPassword string) error
	SearchUsers(ctx context.Context, query string) ([]db.SearchUsersRow, error)
}

type ServiceImpl struct {
	store      repository.Store
	tokenMaker token.Maker
	config     TokenConfig
}

type TokenConfig struct {
	AccessTokenDuration  time.Duration
	RefreshTokenDuration time.Duration
}

func NewService(store repository.Store, tokenMaker token.Maker, config TokenConfig) Service {
	return &ServiceImpl{
		store:      store,
		tokenMaker: tokenMaker,
		config:     config,
	}
}

func (s *ServiceImpl) CreateUser(ctx context.Context, req CreateUserParams) (db.User, error) {
	hashedPassword, err := util.HashPassword(req.Password)
	if err != nil {
		return db.User{}, err
	}

	arg := db.CreateUserParams{
		Phone:        req.Phone,
		Username:     req.Username,
		FullName:     req.FullName,
		PasswordHash: hashedPassword,
	}

	user, err := s.store.CreateUser(ctx, arg)
	if err != nil {
		return db.User{}, err
	}

	return user, nil
}

func (s *ServiceImpl) LoginUser(ctx context.Context, req LoginUserParams) (*LoginUserResult, error) {
	user, err := s.store.GetUserByPhone(ctx, req.Phone)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	err = util.CheckPassword(req.Password, user.PasswordHash)
	if err != nil {
		return nil, errors.New("incorrect password")
	}

	accessToken, accessPayload, err := s.tokenMaker.CreateToken(user.Username, user.ID, s.config.AccessTokenDuration)
	if err != nil {
		return nil, err
	}

	refreshToken, refreshPayload, err := s.tokenMaker.CreateToken(user.Username, user.ID, s.config.RefreshTokenDuration)
	if err != nil {
		return nil, err
	}

	session, err := s.store.CreateSession(ctx, db.CreateSessionParams{
		ID:           refreshPayload.ID,
		UserID:       user.ID,
		RefreshToken: refreshToken,
		UserAgent:    req.UserAgent,
		ClientIp:     req.ClientIP,
		IsBlocked:    false,
		ExpiresAt:    refreshPayload.ExpiredAt,
	})
	if err != nil {
		return nil, err
	}

	return &LoginUserResult{
		SessionID:             session.ID,
		AccessToken:           accessToken,
		AccessTokenExpiresAt:  accessPayload.ExpiredAt,
		RefreshToken:          refreshToken,
		RefreshTokenExpiresAt: refreshPayload.ExpiredAt,
		User:                  user,
	}, nil
}

func (s *ServiceImpl) UpdateEmail(ctx context.Context, req UpdateEmailParams) (db.User, error) {
	_, err := s.store.UpdateUserEmail(ctx, db.UpdateUserEmailParams{
		ID:    req.UserID,
		Email: sql.NullString{String: req.Email, Valid: true},
	})
	if err != nil {
		return db.User{}, err
	}
	// Convert UpdateUserEmailRow to User (manually or by just returning the compatible fields)
	// OR: Helper to fetch full user after update?
	// UpdateUserEmail returns specific fields. The ID should be enough to fetch full user if needed.
	// But UpdateUserEmailRow has most fields. Checking structure...
	// UpdateUserEmailRow has: ID, Phone, PasswordHash, Username, FullName, ...
	// It looks identical to User struct fields but different order/type maybe?
	// Actually generated code UpdateUserEmail returns UpdateUserEmailRow.
	// We should convert it or change interface return type.
	// Let's coerce it to db.User if fields match, or just fetch via ID.
	// Fetching via ID is safer and cleaner architecture-wise.
	return s.store.GetUserByID(ctx, req.UserID)
}

func (s *ServiceImpl) GetUserByID(ctx context.Context, id uuid.UUID) (db.User, error) {
	return s.store.GetUserByID(ctx, id)
}

func (s *ServiceImpl) UpdatePassword(ctx context.Context, userID uuid.UUID, currentPassword, newPassword string) error {
	user, err := s.store.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}

	err = util.CheckPassword(currentPassword, user.PasswordHash)
	if err != nil {
		return errors.New("incorrect current password")
	}

	hashedPassword, err := util.HashPassword(newPassword)
	if err != nil {
		return err
	}

	err = s.store.UpdateUserPassword(ctx, db.UpdateUserPasswordParams{
		ID:           userID,
		PasswordHash: hashedPassword,
	})
	return err
}

func (s *ServiceImpl) SearchUsers(ctx context.Context, query string) ([]db.SearchUsersRow, error) {
	return s.store.SearchUsers(ctx, query)
}
