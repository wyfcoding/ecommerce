package client

import (
	"context"
	"fmt"

	authv1 "ecommerce/api/auth/v1"
	"google.golang.org/grpc"
)

// AuthClient defines the interface to interact with the Auth Service.
type AuthClient interface {
	ValidateToken(ctx context.Context, token string) (isValid bool, userID uint64, username string, err error)
}

type authClient struct {
	client authv1.AuthServiceClient
}

func NewAuthClient(conn *grpc.ClientConn) AuthClient {
	return &authClient{
		client: authv1.NewAuthServiceClient(conn),
	}
}

func (c *authClient) ValidateToken(ctx context.Context, token string) (isValid bool, userID uint64, username string, err error) {
	req := &authv1.ValidateTokenRequest{
		Token: token,
	}
	res, err := c.client.ValidateToken(ctx, req)
	if err != nil {
		return false, 0, "", fmt.Errorf("failed to validate token: %w", err)
	}
	return res.GetIsValid(), res.GetUserId(), res.GetUsername(), nil
}
