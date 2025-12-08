package domain

import (
	"context"
)

// GatewayType 定义支付网关类型
type GatewayType string

const (
	GatewayTypeAlipay GatewayType = "alipay"
	GatewayTypeWechat GatewayType = "wechat"
	GatewayTypeStripe GatewayType = "stripe"
	GatewayTypeMock   GatewayType = "mock"
)

// PaymentRequest 支付请求通用参数
type PaymentGatewayRequest struct {
	OrderID     string            // 订单号
	Amount      int64             // 金额（分）
	Currency    string            // 货币
	Description string            // 描述
	ClientIP    string            // 客户端IP
	ReturnURL   string            // 前端跳转地址
	NotifyURL   string            // 后端回调地址
	ExtraData   map[string]string // 额外数据
}

// PaymentResponse 支付响应通用参数
type PaymentGatewayResponse struct {
	TransactionID string // 网关交易ID
	PaymentURL    string // 支付跳转链接 (Web/Wap)
	QRCode        string // 二维码链接 (Scan)
	AppParam      string // App支付参数 (SDK)
	RawResponse   string // 原始响应
}

// RefundRequest 退款请求通用参数
type RefundGatewayRequest struct {
	PaymentID     string // 支付ID (我们系统的)
	TransactionID string // 网关交易ID
	RefundID      string // 退款ID (我们系统的)
	Amount        int64  // 退款金额
	Reason        string // 原因
}

// RefundResponse 退款响应通用参数
type RefundGatewayResponse struct {
	RefundID    string // 网关退款ID
	Status      string // 网关状态
	RawResponse string // 原始响应
}

// PaymentGateway 支付网关抽象接口
type PaymentGateway interface {
	// Pay 发起支付
	Pay(ctx context.Context, req *PaymentGatewayRequest) (*PaymentGatewayResponse, error)

	// Query 查询支付状态
	Query(ctx context.Context, transactionID string) (*PaymentGatewayResponse, error)

	// Refund 发起退款
	Refund(ctx context.Context, req *RefundGatewayRequest) (*RefundGatewayResponse, error)

	// QueryRefund 查询退款状态
	QueryRefund(ctx context.Context, refundID string) (*RefundGatewayResponse, error)

	// VerifyCallback 验证回调签名
	VerifyCallback(ctx context.Context, data map[string]string) (bool, error)

	// GetType 获取网关类型
	GetType() GatewayType
}
