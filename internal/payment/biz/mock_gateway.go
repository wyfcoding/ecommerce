
package biz

import (
	"context"
	"fmt"
	"time"
)

// MockGateway 是一个模拟外部支付网关的客户端。	ype MockGateway struct {
}

// NewMockGateway 创建一个新的模拟网关客户端。
func NewMockGateway() *MockGateway {
	return &MockGateway{}
}

// PaymentGatewayClient 定义了支付网关客户端需要实现的接口。
type PaymentGatewayClient interface {
	CreatePayment(ctx context.Context, amount float64, orderID string) (string, string, error)
	CreateRefund(ctx context.Context, transactionID string, amount float64) (string, error)
}

// CreatePayment 模拟创建一个支付请求。
// 在真实场景中，它会调用像 Stripe 或 Alipay 的 API。
func (m *MockGateway) CreatePayment(ctx context.Context, amount float64, orderID string) (string, string, error) {
	// 生成一个假的交易ID
	transactionID := fmt.Sprintf("mock_txn_%d", time.Now().UnixNano())
	// 生成一个假的支付URL
	paymentURL := fmt.Sprintf("https://mock-payment-gateway.com/pay?order_id=%s&amount=%.2f", orderID, amount)
	return transactionID, paymentURL, nil
}

// CreateRefund 模拟创建一个退款请求。
func (m *MockGateway) CreateRefund(ctx context.Context, transactionID string, amount float64) (string, error) {
	// 生成一个假的退款ID
	refundID := fmt.Sprintf("mock_refund_%d", time.Now().UnixNano())
	return refundID, nil
}
