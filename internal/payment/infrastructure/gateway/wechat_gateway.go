package gateway

import (
	"context"
	"time"

	"github.com/wyfcoding/ecommerce/internal/payment/domain"
)

type WechatGateway struct{}

func NewWechatGateway() *WechatGateway { return &WechatGateway{} }

func (g *WechatGateway) PreAuth(ctx context.Context, req *domain.PaymentGatewayRequest) (*domain.PaymentGatewayResponse, error) {
	return &domain.PaymentGatewayResponse{
		TransactionID: "WX_AUTH_" + req.OrderID,
		PaymentURL:    "weixin://wxpay/bizpayurl",
	}, nil
}

func (g *WechatGateway) Capture(ctx context.Context, transactionID string, amount int64) (*domain.PaymentGatewayResponse, error) {
	return &domain.PaymentGatewayResponse{
		TransactionID: transactionID,
	}, nil
}

func (g *WechatGateway) Void(ctx context.Context, transactionID string) error {
	return nil
}

func (g *WechatGateway) Refund(ctx context.Context, transactionID string, amount int64) error {
	return nil
}

func (g *WechatGateway) DownloadBill(ctx context.Context, date time.Time) ([]*domain.GatewayBillItem, error) {
	// TODO: 对接微信支付账单下载 API
	return []*domain.GatewayBillItem{}, nil
}
