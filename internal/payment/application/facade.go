package application

import (
	"context"

	"github.com/wyfcoding/ecommerce/internal/payment/domain"
)

// PaymentFacade 支付门面服务。
type PaymentFacade struct {
	command *PaymentCommandService
	query   *PaymentQuery
}

// NewPaymentFacade 构造函数。
func NewPaymentFacade(command *PaymentCommandService, query *PaymentQuery) *PaymentFacade {
	return &PaymentFacade{
		command: command,
		query:   query,
	}
}

// InitiatePayment 发起支付。
func (f *PaymentFacade) InitiatePayment(ctx context.Context, cmd *InitiatePaymentCommand) (*domain.Payment, *domain.PaymentGatewayResponse, error) {
	return f.command.InitiatePayment(ctx, cmd)
}

// CapturePayment 捕获支付。
func (f *PaymentFacade) CapturePayment(ctx context.Context, cmd *CapturePaymentCommand) error {
	return f.command.CapturePayment(ctx, cmd)
}

// RequestRefund 请求退款。
func (f *PaymentFacade) RequestRefund(ctx context.Context, cmd *RefundPaymentCommand) (*domain.Refund, error) {
	return f.command.RequestRefund(ctx, cmd)
}

// HandlePaymentCallback 处理支付回调。
func (f *PaymentFacade) HandlePaymentCallback(ctx context.Context, userID uint64, paymentNo string, success bool, transactionID, thirdPartyNo string, callbackData map[string]string) error {
	return f.command.HandlePaymentCallback(ctx, userID, paymentNo, success, transactionID, thirdPartyNo, callbackData)
}

// GetPayment 获取支付详情。
func (f *PaymentFacade) GetPayment(ctx context.Context, userID uint64, paymentID uint64) (*domain.Payment, error) {
	return f.query.GetPayment(ctx, userID, paymentID)
}

// GetPaymentByOrder 获取订单支付信息。
func (f *PaymentFacade) GetPaymentByOrder(ctx context.Context, userID uint64, orderID uint64) (*domain.Payment, error) {
	return f.query.GetPaymentByOrder(ctx, userID, orderID)
}
