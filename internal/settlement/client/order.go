package client

import (
	"context"
	"fmt"

	orderv1 "ecommerce/api/order/v1"
	"google.golang.org/grpc"
)

// OrderClient defines the interface to interact with the Order Service (to get order details).
type OrderClient interface {
	GetOrderDetails(ctx context.Context, orderID uint64) (*orderv1.OrderInfo, error)
}

type orderClient struct {
	client orderv1.OrderClient
}

func NewOrderClient(conn *grpc.ClientConn) OrderClient {
	return &orderClient{
		client: orderv1.NewOrderClient(conn),
	}
}

func (c *orderClient) GetOrderDetails(ctx context.Context, orderID uint64) (*orderv1.OrderInfo, error) {
	req := &orderv1.GetOrderDetailRequest{OrderId: fmt.Sprintf("%d", orderID)}
	res, err := c.client.GetOrderDetail(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get order details: %w", err)
	}
	return res.GetOrder(), nil
}
