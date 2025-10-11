package biz

import (
	"context"
	"errors"
	"time"

	"ecommerce/pkg/hash" // Import the hash package
	"ecommerce/pkg/jwt"
)

var (
	ErrInvalidCredentials = errors.New("invalid username or password")
	ErrInvalidToken       = errors.New("invalid token")
)

// Session represents a user session.	ype Session struct {
	ID        uint
	UserID    uint64
	Token     string
	ExpiresAt time.Time
}

// AuthRepo defines the interface for authentication data access.
type AuthRepo interface {
	CreateSession(ctx context.Context, session *Session) (*Session, error)
	GetSessionByToken(ctx context.Context, token string) (*Session, error)
}

// UserClient defines the interface to interact with the User Service.
type UserClient interface {
	VerifyPassword(ctx context.Context, username, password string) (bool, *User, error)
}

// User represents the user model from the user service.
type User struct {
    ID       uint64
    Username string
}


// AuthUsecase is the business logic for authentication.
type AuthUsecase struct {
	repo       AuthRepo
	userClient UserClient
	jwtSecret  string
	jwtIssuer  string
	jwtExpire  time.Duration
}

// NewAuthUsecase creates a new AuthUsecase.
func NewAuthUsecase(repo AuthRepo, userClient UserClient, jwtSecret, jwtIssuer string, jwtExpire time.Duration) *AuthUsecase {
	return &AuthUsecase{
		repo:       repo,
		userClient: userClient,
		jwtSecret:  jwtSecret,
		jwtIssuer:  jwtIssuer,
		jwtExpire:  jwtExpire,
	}
}

// Login handles user login and token issuance.
func (uc *AuthUsecase) Login(ctx context.Context, username, password string) (string, error) {
	// 1. Verify credentials using User Service
	valid, user, err := uc.userClient.VerifyPassword(ctx, username, password)
	if err != nil {
		return "", err
	}
	if !valid {
		return "", ErrInvalidCredentials
	}

	// 2. Generate Access Token
	accessToken, err := jwt.GenerateToken(user.ID, user.Username, uc.jwtSecret, uc.jwtIssuer, uc.jwtExpire)
	if err != nil {
		return "", err
	}

	// 3. Save session (optional, but good for revocation)
	// ... implementation for session saving ...

	return accessToken, nil
}

// ValidateToken validates an access token.
func (uc *AuthUsecase) ValidateToken(ctx context.Context, token string) (bool, uint64, string, error) {
	claims, err := jwt.ParseToken(token, uc.jwtSecret)
	if err != nil {
		return false, 0, "", ErrInvalidToken
	}

	// Optional: Check if session exists in repo for revocation
	// ... implementation for session checking ...

	return true, claims.UserID, claims.Username, nil
}