package gateway

import (
	"context"
	"fmt"

	"github.com/wyfcoding/ecommerce/internal/payment/domain"
)

// StripeGateway Stripe网关模拟实现
type StripeGateway struct{}

func NewStripeGateway() *StripeGateway {
	return &StripeGateway{}
}

func (g *StripeGateway) Pay(ctx context.Context, req *domain.PaymentGatewayRequest) (*domain.PaymentGatewayResponse, error) {
	return &domain.PaymentGatewayResponse{
		TransactionID: "pi_fake_123456",
		PaymentURL:    "https://checkout.stripe.com/pay/...",
		RawResponse:   `{"id": "pi_123"}`,
	}, nil
}

func (g *StripeGateway) Query(ctx context.Context, transactionID string) (*domain.PaymentGatewayResponse, error) {
	return &domain.PaymentGatewayResponse{
		TransactionID: transactionID,
		RawResponse:   `{"status": "succeeded"}`,
	}, nil
}

func (g *StripeGateway) Refund(ctx context.Context, req *domain.RefundGatewayRequest) (*domain.RefundGatewayResponse, error) {
	return &domain.RefundGatewayResponse{
		RefundID:    "re_fake_123",
		Status:      "succeeded",
		RawResponse: `{"status": "succeeded"}`,
	}, nil
}

func (g *StripeGateway) QueryRefund(ctx context.Context, refundID string) (*domain.RefundGatewayResponse, error) {
	return nil, fmt.Errorf("not implemented")
}

func (g *StripeGateway) VerifyCallback(ctx context.Context, data map[string]string) (bool, error) {
	return true, nil
}

func (g *StripeGateway) GetType() domain.GatewayType {
	return domain.GatewayTypeStripe
}
