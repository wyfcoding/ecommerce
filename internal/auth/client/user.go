package client

import (
	"context"

	"ecommerce/internal/auth/model"
)

// UserClient defines the interface to interact with the User Service.
type UserClient interface {
	VerifyPassword(ctx context.Context, username, password string) (bool, *model.User, error)
}
