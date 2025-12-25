package application

import (
	"context"
	"log/slog"

	settlementv1 "github.com/wyfcoding/ecommerce/goapi/settlement/v1"
	"github.com/wyfcoding/ecommerce/internal/payment/domain"
)

// Payment 支付应用服务门面
type Payment struct {
	Processor       *PaymentProcessor
	CallbackHandler *CallbackHandler
	RefundService   *RefundService
	Query           *PaymentQuery
	settlementCli   settlementv1.SettlementClient
	logger          *slog.Logger
}

// NewPayment 创建支付服务实例。
func NewPayment(
	processor *PaymentProcessor,
	callbackHandler *CallbackHandler,
	refundService *RefundService,
	query *PaymentQuery,
	settlementCli settlementv1.SettlementClient,
	logger *slog.Logger,
) *Payment {
	return &Payment{
		Processor:       processor,
		CallbackHandler: callbackHandler,
		RefundService:   refundService,
		Query:           query,
		settlementCli:   settlementCli,
		logger:          logger,
	}
}

// InitiatePayment 发起支付交易
func (s *Payment) InitiatePayment(ctx context.Context, orderID uint64, userID uint64, amount int64, paymentMethodStr string) (*domain.Payment, *domain.PaymentGatewayResponse, error) {
	return s.Processor.InitiatePayment(ctx, orderID, userID, amount, paymentMethodStr)
}

// HandlePaymentCallback 处理支付结果异步回调
func (s *Payment) HandlePaymentCallback(ctx context.Context, paymentNo string, success bool, transactionID, thirdPartyNo string, callbackData map[string]string) error {
	return s.CallbackHandler.HandlePaymentCallback(ctx, paymentNo, success, transactionID, thirdPartyNo, callbackData)
}

// RequestRefund 发起退款申请
func (s *Payment) RequestRefund(ctx context.Context, paymentID uint64, amount int64, reason string) (*domain.Refund, error) {
	return s.RefundService.RequestRefund(ctx, paymentID, amount, reason)
}

// GetPaymentStatus 获取支付详情
func (s *Payment) GetPaymentStatus(ctx context.Context, id uint64) (*domain.Payment, error) {
	return s.Query.GetPaymentStatus(ctx, id)
}
