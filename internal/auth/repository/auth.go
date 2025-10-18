package repository

import (
	"context"

	"ecommerce/internal/auth/model"
)

// AuthRepo defines the interface for authentication data access.
type AuthRepo interface {
	CreateSession(ctx context.Context, session *model.Session) (*model.Session, error)
	GetSessionByToken(ctx context.Context, token string) (*model.Session, error)
}