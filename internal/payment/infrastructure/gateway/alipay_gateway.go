package gateway

import (
	"context"
	"fmt"

	"github.com/wyfcoding/ecommerce/internal/payment/domain"
)

// AlipayGateway 支付宝网关模拟实现
type AlipayGateway struct{}

func NewAlipayGateway() *AlipayGateway {
	return &AlipayGateway{}
}

func (g *AlipayGateway) Pay(ctx context.Context, req *domain.PaymentGatewayRequest) (*domain.PaymentGatewayResponse, error) {
	// 模拟支付宝支付流程
	// 实际场景下会生成一个 form 表单或获取 SDK 参数
	return &domain.PaymentGatewayResponse{
		TransactionID: "", // 支付宝通常是回调时才给交易号，这里模拟为空
		PaymentURL:    "https://mock.alipay.com/gateway.do?params=...",
		RawResponse:   `{"code":"10000", "msg":"Success"}`,
	}, nil
}

func (g *AlipayGateway) Query(ctx context.Context, transactionID string) (*domain.PaymentGatewayResponse, error) {
	return &domain.PaymentGatewayResponse{
		TransactionID: transactionID,
		RawResponse:   `{"status":"TRADE_SUCCESS"}`,
	}, nil
}

func (g *AlipayGateway) Refund(ctx context.Context, req *domain.RefundGatewayRequest) (*domain.RefundGatewayResponse, error) {
	return &domain.RefundGatewayResponse{
		RefundID:    "REF_" + req.TransactionID,
		Status:      "SUCCESS",
		RawResponse: `{"code":"10000", "msg":"Success"}`,
	}, nil
}

func (g *AlipayGateway) QueryRefund(ctx context.Context, refundID string) (*domain.RefundGatewayResponse, error) {
	return nil, fmt.Errorf("not implemented")
}

func (g *AlipayGateway) VerifyCallback(ctx context.Context, data map[string]string) (bool, error) {
	// 模拟验签：总是通过
	return true, nil
}

func (g *AlipayGateway) GetType() domain.GatewayType {
	return domain.GatewayTypeAlipay
}
