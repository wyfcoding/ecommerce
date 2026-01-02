package gateway

import (
	"context"
	"time"

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

func (g *AlipayGateway) DownloadBill(ctx context.Context, date time.Time) ([]*domain.GatewayBillItem, error) {
	// TODO: 对接支付宝对账单拉取 API (如 alipay.data.bill.balance.query)
	return []*domain.GatewayBillItem{}, nil
}
