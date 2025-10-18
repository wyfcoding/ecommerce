package client

import (
	"context"
	"fmt"

	orderv1 "ecommerce/api/order/v1"
	"google.golang.org/grpc"
)

// OrderServiceClient defines the interface for interacting with the Order Service.
type OrderServiceClient interface {
	CreateOrderForFlashSale(ctx context.Context, userID, productID string, quantity int32, price float64) (string, error)
	CompensateCreateOrder(ctx context.Context, orderID string) error
}

type orderServiceClient struct {
	client orderv1.OrderClient
}

func NewOrderServiceClient(conn *grpc.ClientConn) OrderServiceClient {
	return &orderServiceClient{
		client: orderv1.NewOrderClient(conn),
	}
}

func (c *orderServiceClient) CreateOrderForFlashSale(ctx context.Context, userID, productID string, quantity int32, price float64) (string, error) {
	req := &orderv1.CreateOrderRequest{
		UserId:    userID,
		ProductId: productID,
		Quantity:  quantity,
		Price:     price,
		OrderType: "flash_sale",
	}
	res, err := c.client.CreateOrder(ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to create order: %w", err)
	}
	return res.GetOrderId(), nil
}

func (c *orderServiceClient) CompensateCreateOrder(ctx context.Context, orderID string) error {
	req := &orderv1.CompensateCreateOrderRequest{
		OrderId: orderID,
	}
	_, err := c.client.CompensateCreateOrder(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to compensate create order: %w", err)
	}
	return nil
}
