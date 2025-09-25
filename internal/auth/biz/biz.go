package biz

import (
	"context"
	"errors"
	"time"

	"ecommerce/pkg/jwt" // Assuming pkg/jwt exists
)

var (
	ErrInvalidCredentials = errors.New("invalid username or password")
	ErrInvalidToken       = errors.New("invalid token")
	ErrUserNotFound       = errors.New("user not found")
)

// Session represents a user session in the business logic layer.
type Session struct {
	ID        uint
	UserID    uint64
	Token     string
	ExpiresAt time.Time
}

// RefreshToken represents a refresh token in the business logic layer.
type RefreshToken struct {
	ID        uint
	UserID    uint64
	Token     string
	ExpiresAt time.Time
}

// AuthRepo defines the interface for authentication data access.
type AuthRepo interface {
	CreateSession(ctx context.Context, session *Session) (*Session, error)
	GetSessionByToken(ctx context.Context, token string) (*Session, error)
	DeleteSessionByToken(ctx context.Context, token string) error
	CreateRefreshToken(ctx context.Context, refreshToken *RefreshToken) (*RefreshToken, error)
	GetRefreshTokenByToken(ctx context.Context, token string) (*RefreshToken, error)
	DeleteRefreshTokenByToken(ctx context.Context, token string) error
}

// UserInfo represents basic user info from User Service.
type UserInfo struct {
	ID       uint64
	Username string
	Password string // Hashed password
}

// UserClient defines the interface to interact with the User Service.
type UserClient interface {
	GetUserByUsername(ctx context.Context, username string) (*UserInfo, error)
	// TODO: Add method to verify password hash
}

// AuthUsecase is the business logic for authentication.
type AuthUsecase struct {
	repo       AuthRepo
	userClient UserClient
	jwtSecret  []byte
	// TODO: Add password hasher
}

// NewAuthUsecase creates a new AuthUsecase.
func NewAuthUsecase(repo AuthRepo, userClient UserClient, jwtSecret []byte) *AuthUsecase {
	return &AuthUsecase{
		repo:       repo,
		userClient: userClient,
		jwtSecret:  jwtSecret,
	}
}

// Login handles user login and token issuance.
func (uc *AuthUsecase) Login(ctx context.Context, username, password string) (accessToken, refreshToken string, expiresIn int64, err error) {
	// 1. Get user from User Service
	user, err := uc.userClient.GetUserByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) { // Assuming UserClient returns ErrUserNotFound
			return "", "", 0, ErrInvalidCredentials
		}
		return "", "", 0, err
	}
	if user == nil {
		return "", "", 0, ErrInvalidCredentials
	}

	// 2. Verify password (TODO: Use a proper password hasher)
	if user.Password != password { // Placeholder for password verification
		return "", "", 0, ErrInvalidCredentials
	}

	// 3. Generate Access Token
	accessToken, expiresIn, err = jwt.GenerateToken(user.ID, user.Username, uc.jwtSecret, 15*time.Minute) // Access token valid for 15 minutes
	if err != nil {
		return "", "", 0, err
	}

	// 4. Generate Refresh Token
	refreshToken, _, err = jwt.GenerateToken(user.ID, user.Username, uc.jwtSecret, 7*24*time.Hour) // Refresh token valid for 7 days
	if err != nil {
		return "", "", 0, err
	}

	// 5. Save session and refresh token (optional, but good for revocation)
	_, err = uc.repo.CreateSession(ctx, &Session{
		UserID:    user.ID,
		Token:     accessToken,
		ExpiresAt: time.Now().Add(15 * time.Minute),
	})
	if err != nil {
		return "", "", 0, err
	}

	_, err = uc.repo.CreateRefreshToken(ctx, &RefreshToken{
		UserID:    user.ID,
		Token:     refreshToken,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	})
	if err != nil {
		return "", "", 0, err
	}

	return accessToken, refreshToken, expiresIn, nil
}

// ValidateToken validates an access token.
func (uc *AuthUsecase) ValidateToken(ctx context.Context, token string) (isValid bool, userID uint64, username string, err error) {
	claims, err := jwt.ParseToken(token, uc.jwtSecret)
	if err != nil {
		return false, 0, "", ErrInvalidToken
	}

	// Check if session exists (for revocation)
	session, err := uc.repo.GetSessionByToken(ctx, token)
	if err != nil {
		return false, 0, "", err
	}
	if session == nil {
		return false, 0, "", ErrInvalidToken // Token revoked or not found
	}

	return true, claims.UserID, claims.Username, nil
}
