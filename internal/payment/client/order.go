package client

import (
	"context"
	"fmt"

	orderv1 "ecommerce/api/order/v1"
	paymentv1 "ecommerce/api/payment/v1"
	"google.golang.org/grpc"
)

// OrderClient 定义了与订单服务交互的接口。
type OrderClient interface {
	// 这里可以定义从订单服务获取信息的方法
	// GetOrderInfo(ctx context.Context, orderID string) (*Order, error)
	ProcessPaymentNotification(ctx context.Context, orderID uint64, paymentStatus paymentv1.PaymentStatus) error
}

type orderClient struct {
	client orderv1.OrderClient
}

func NewOrderClient(conn *grpc.ClientConn) OrderClient {
	return &orderClient{
		client: orderv1.NewOrderClient(conn),
	}
}

func (c *orderClient) ProcessPaymentNotification(ctx context.Context, orderID uint64, paymentStatus paymentv1.PaymentStatus) error {
	req := &orderv1.ProcessPaymentNotificationRequest{
		OrderId:       orderID,
		PaymentStatus: paymentStatus,
	}
	_, err := c.client.ProcessPaymentNotification(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to process payment notification in order service: %w", err)
	}
	return nil
}
