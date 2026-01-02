package application

import (
	"context"
	"log/slog"

	settlementv1 "github.com/wyfcoding/ecommerce/goapi/settlement/v1"
	"github.com/wyfcoding/ecommerce/internal/payment/domain"
)

// PaymentService 支付应用服务门面
type PaymentService struct {
	Processor       *PaymentProcessor
	CallbackHandler *CallbackHandler
	RefundService   *RefundService
	Query           *PaymentQuery
	settlementCli   settlementv1.SettlementServiceClient
	logger          *slog.Logger
}

// NewPaymentService 创建支付服务实例。
func NewPaymentService(
	processor *PaymentProcessor,
	callbackHandler *CallbackHandler,
	refundService *RefundService,
	query *PaymentQuery,
	settlementCli settlementv1.SettlementServiceClient,
	logger *slog.Logger,
) *PaymentService {
	return &PaymentService{
		Processor:       processor,
		CallbackHandler: callbackHandler,
		RefundService:   refundService,
		Query:           query,
		settlementCli:   settlementCli,
		logger:          logger,
	}
}

func (s *PaymentService) InitiatePayment(ctx context.Context, orderID uint64, userID uint64, amount int64, paymentMethod string) (*domain.Payment, *domain.PaymentGatewayResponse, error) {
	return s.Processor.InitiatePayment(ctx, orderID, userID, amount, paymentMethod)
}

func (s *PaymentService) HandlePaymentCallback(ctx context.Context, userID uint64, paymentNo string, success bool, transactionID, thirdPartyNo string, callbackData map[string]string) error {
	return s.CallbackHandler.HandlePaymentCallback(ctx, userID, paymentNo, success, transactionID, thirdPartyNo, callbackData)
}

func (s *PaymentService) RequestRefund(ctx context.Context, userID, id uint64, amount int64, reason string) (*domain.Refund, error) {
	return s.RefundService.RequestRefund(ctx, userID, id, amount, reason)
}

func (s *PaymentService) CapturePayment(ctx context.Context, userID uint64, paymentNo string, amount int64) error {
	return s.Processor.CapturePayment(ctx, userID, paymentNo, amount)
}

func (s *PaymentService) GetPaymentStatus(ctx context.Context, userID, id uint64) (*domain.Payment, error) {
	return s.Query.GetPaymentStatus(ctx, userID, id)
}

func (s *PaymentService) GetUserIDByPaymentNo(ctx context.Context, paymentNo string) (uint64, error) {
	return s.Processor.paymentRepo.GetUserIDByPaymentNo(ctx, paymentNo)
}
