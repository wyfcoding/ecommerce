package gateway

import (
	"context"

	"github.com/wyfcoding/ecommerce/internal/payment/domain"
)

type MockGateway struct{}

func NewMockGateway() *MockGateway { return &MockGateway{} }

func (g *MockGateway) PreAuth(ctx context.Context, req *domain.PaymentGatewayRequest) (*domain.PaymentGatewayResponse, error) {
	return &domain.PaymentGatewayResponse{
		TransactionID: "MOCK_AUTH_" + req.OrderID,
		PaymentURL:    "http://mock.gateway/pay",
	}, nil
}

func (g *MockGateway) Capture(ctx context.Context, transactionID string, amount int64) (*domain.PaymentGatewayResponse, error) {
	return &domain.PaymentGatewayResponse{
		TransactionID: transactionID,
	}, nil
}

func (g *MockGateway) Void(ctx context.Context, transactionID string) error {
	return nil
}

func (g *MockGateway) Refund(ctx context.Context, transactionID string, amount int64) error {
	return nil
}
