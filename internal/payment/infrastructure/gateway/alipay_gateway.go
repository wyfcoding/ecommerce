package gateway

import (
	"context"

	"github.com/wyfcoding/ecommerce/internal/payment/domain"
)

type AlipayGateway struct{}

func NewAlipayGateway() *AlipayGateway { return &AlipayGateway{} }

func (g *AlipayGateway) PreAuth(ctx context.Context, req *domain.PaymentGatewayRequest) (*domain.PaymentGatewayResponse, error) {
	return &domain.PaymentGatewayResponse{
		TransactionID: "ALI_AUTH_" + req.OrderID,
		PaymentURL:    "https://mock.alipay.com/auth",
	}, nil
}

func (g *AlipayGateway) Capture(ctx context.Context, transactionID string, amount int64) (*domain.PaymentGatewayResponse, error) {
	return &domain.PaymentGatewayResponse{
		TransactionID: transactionID,
	}, nil
}

func (g *AlipayGateway) Void(ctx context.Context, transactionID string) error {
	return nil
}

func (g *AlipayGateway) Refund(ctx context.Context, transactionID string, amount int64) error {
	return nil
}
