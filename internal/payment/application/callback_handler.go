package application

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/wyfcoding/ecommerce/internal/payment/domain"
	"github.com/wyfcoding/pkg/lock"
)

type CallbackHandler struct {
	paymentRepo domain.PaymentRepository
	gateways    map[domain.GatewayType]domain.PaymentGateway
	lockSvc     *lock.RedisLock
	logger      *slog.Logger
}

func NewCallbackHandler(
	paymentRepo domain.PaymentRepository,
	gateways map[domain.GatewayType]domain.PaymentGateway,
	lockSvc *lock.RedisLock,
	logger *slog.Logger,
) *CallbackHandler {
	return &CallbackHandler{
		paymentRepo: paymentRepo,
		gateways:    gateways,
		lockSvc:     lockSvc,
		logger:      logger,
	}
}

// HandlePaymentCallback 这里通常用于非 Pre-auth 的直接支付流，或者异步通知
func (s *CallbackHandler) HandlePaymentCallback(ctx context.Context, paymentNo string, success bool, transactionID, thirdPartyNo string, callbackData map[string]string) error {
	s.logger.InfoContext(ctx, "handling payment callback", "payment_no", paymentNo, "success", success)

	lockKey := fmt.Sprintf("lock:payment:callback:%s", paymentNo)
	token, err := s.lockSvc.Lock(ctx, lockKey, 10*time.Second)
	if err != nil {
		return err
	}
	defer s.lockSvc.Unlock(ctx, lockKey, token)

	payment, err := s.paymentRepo.FindByPaymentNo(ctx, paymentNo)
	if err != nil || payment == nil {
		return fmt.Errorf("payment not found")
	}

	// 状态机处理
	event := "PAY_DIRECT"
	if !success {
		// 简化处理为失败
		return fmt.Errorf("payment failed from gateway")
	}

	if err := payment.Trigger(ctx, event, "Callback received"); err != nil {
		return err
	}
	payment.TransactionID = transactionID
	payment.ThirdPartyNo = thirdPartyNo
	now := time.Now()
	payment.PaidAt = &now

	return s.paymentRepo.Update(ctx, payment)
}
