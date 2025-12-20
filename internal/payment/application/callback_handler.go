package application

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/wyfcoding/ecommerce/internal/payment/domain"
)

// CallbackHandler 处理 HTTP 或 gRPC 请求。
type CallbackHandler struct {
	paymentRepo domain.PaymentRepository
	gateways    map[domain.GatewayType]domain.PaymentGateway
	logger      *slog.Logger
}

// NewCallbackHandler 处理 HTTP 或 gRPC 请求。
func NewCallbackHandler(
	paymentRepo domain.PaymentRepository,
	gateways map[domain.GatewayType]domain.PaymentGateway,
	logger *slog.Logger,
) *CallbackHandler {
	return &CallbackHandler{
		paymentRepo: paymentRepo,
		gateways:    gateways,
		logger:      logger,
	}
}

// HandlePaymentCallback 处理支付结果异步回调
func (s *CallbackHandler) HandlePaymentCallback(ctx context.Context, paymentNo string, success bool, transactionID, thirdPartyNo string, callbackData map[string]string) error {
	s.logger.InfoContext(ctx, "handling payment callback", "payment_no", paymentNo, "success", success)

	payment, err := s.paymentRepo.FindByPaymentNo(ctx, paymentNo)
	if err != nil || payment == nil {
		s.logger.ErrorContext(ctx, "payment not found", "payment_no", paymentNo)
		return fmt.Errorf("payment not found")
	}

	// 验证签名
	gateway, ok := s.gateways[payment.GatewayType]
	if ok {
		valid, err := gateway.VerifyCallback(ctx, callbackData)
		if err != nil || !valid {
			s.logger.WarnContext(ctx, "invalid callback signature", "payment_no", paymentNo)
			return fmt.Errorf("invalid signature")
		}
	}

	// 领域层处理状态变更
	if err := payment.Process(success, transactionID, thirdPartyNo); err != nil {
		s.logger.ErrorContext(ctx, "failed to process state transition", "error", err)
		return err
	}

	if err := s.paymentRepo.Update(ctx, payment); err != nil {
		return err
	}

	return nil
}
