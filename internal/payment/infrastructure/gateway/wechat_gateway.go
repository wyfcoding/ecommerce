package gateway

import (
	"context"
	"fmt"

	"github.com/wyfcoding/ecommerce/internal/payment/domain"
)

// WechatGateway 微信支付网关模拟实现
type WechatGateway struct{}

// NewWechatGateway 函数。
func NewWechatGateway() *WechatGateway {
	return &WechatGateway{}
}

func (g *WechatGateway) Pay(ctx context.Context, req *domain.PaymentGatewayRequest) (*domain.PaymentGatewayResponse, error) {
	return &domain.PaymentGatewayResponse{
		TransactionID: "WX_" + req.OrderID, // 微信预支付ID
		PaymentURL:    "weixin://wxpay/bizpayurl?pr=...",
		RawResponse:   `<xml><return_code>SUCCESS</return_code></xml>`,
	}, nil
}

func (g *WechatGateway) Query(ctx context.Context, transactionID string) (*domain.PaymentGatewayResponse, error) {
	return &domain.PaymentGatewayResponse{
		TransactionID: transactionID,
		RawResponse:   `<xml>...</xml>`,
	}, nil
}

func (g *WechatGateway) Refund(ctx context.Context, req *domain.RefundGatewayRequest) (*domain.RefundGatewayResponse, error) {
	return &domain.RefundGatewayResponse{
		RefundID:    "REF_" + req.TransactionID,
		Status:      "SUCCESS",
		RawResponse: `<xml>...</xml>`,
	}, nil
}

func (g *WechatGateway) QueryRefund(ctx context.Context, refundID string) (*domain.RefundGatewayResponse, error) {
	return nil, fmt.Errorf("not implemented")
}

func (g *WechatGateway) VerifyCallback(ctx context.Context, data map[string]string) (bool, error) {
	return true, nil
}

func (g *WechatGateway) GetType() domain.GatewayType {
	return domain.GatewayTypeWechat
}
