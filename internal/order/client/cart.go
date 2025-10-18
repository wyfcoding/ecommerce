package client

import (
	"context"
	"fmt"

	cartv1 "ecommerce/api/cart/v1"
	"google.golang.org/grpc"
)

// CartClient 定义了与购物车服务交互的接口
type CartClient interface {
	ClearCart(ctx context.Context, userID uint64) error
}

type cartClient struct {
	client cartv1.CartClient
}

func NewCartClient(conn *grpc.ClientConn) CartClient {
	return &cartClient{
		client: cartv1.NewCartClient(conn),
	}
}

func (c *cartClient) ClearCart(ctx context.Context, userID uint64) error {
	req := &cartv1.ClearCartRequest{UserId: userID}
	_, err := c.client.ClearCart(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to clear cart: %w", err)
	}
	return nil
}
