package client

import (
	"context"
	"fmt"

	paymentv1 "ecommerce/api/payment/v1"
	"ecommerce/internal/order/model"
	"google.golang.org/grpc"
)

// PaymentClient 定义了与支付服务交互的接口
type PaymentClient interface {
	CreatePayment(ctx context.Context, userID uint64, orderID string, amount float64) (*model.PaymentInfo, error)
	ProcessPaymentNotification(ctx context.Context, paymentID string, data map[string]string) error
}

type paymentClient struct {
	client paymentv1.PaymentClient
}

func NewPaymentClient(conn *grpc.ClientConn) PaymentClient {
	return &paymentClient{
		client: paymentv1.NewPaymentClient(conn),
	}
}

func (c *paymentClient) CreatePayment(ctx context.Context, userID uint64, orderID string, amount float64) (*model.PaymentInfo, error) {
	req := &paymentv1.CreatePaymentRequest{
		UserId:  userID,
		OrderId: orderID,
		Amount:  amount,
	}
	res, err := c.client.CreatePayment(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create payment: %w", err)
	}
	return &model.PaymentInfo{PaymentID: res.GetPaymentId(), PaymentURL: res.GetPaymentUrl()}, nil
}

func (c *paymentClient) ProcessPaymentNotification(ctx context.Context, paymentID string, data map[string]string) error {
	req := &paymentv1.ProcessPaymentNotificationRequest{
		PaymentId: paymentID,
		Data:      data,
	}
	_, err := c.client.ProcessPaymentNotification(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to process payment notification: %w", err)
	}
	return nil
}
