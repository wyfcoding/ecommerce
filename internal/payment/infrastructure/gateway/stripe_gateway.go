package gateway

import (
	"context"
	"time"

	"github.com/wyfcoding/ecommerce/internal/payment/domain"
)

type StripeGateway struct{}

func NewStripeGateway() *StripeGateway { return &StripeGateway{} }

func (g *StripeGateway) PreAuth(ctx context.Context, req *domain.PaymentGatewayRequest) (*domain.PaymentGatewayResponse, error) {
	return &domain.PaymentGatewayResponse{
		TransactionID: "pi_stripe_auth_" + req.OrderID,
		PaymentURL:    "https://checkout.stripe.com/pay",
	}, nil
}

func (g *StripeGateway) Capture(ctx context.Context, transactionID string, amount int64) (*domain.PaymentGatewayResponse, error) {
	return &domain.PaymentGatewayResponse{
		TransactionID: transactionID,
	}, nil
}

func (g *StripeGateway) Void(ctx context.Context, transactionID string) error {
	return nil
}

func (g *StripeGateway) Refund(ctx context.Context, transactionID string, amount int64) error {
	return nil
}

func (g *StripeGateway) DownloadBill(ctx context.Context, date time.Time) ([]*domain.GatewayBillItem, error) {
	// TODO: 对接 Stripe 报表/对账单 API
	return []*domain.GatewayBillItem{}, nil
}
