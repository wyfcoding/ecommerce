package application

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/wyfcoding/ecommerce/internal/payment/domain"
	"github.com/wyfcoding/pkg/lock"
)

// CallbackHandler 处理 HTTP 或 gRPC 请求。
type CallbackHandler struct {
	paymentRepo domain.PaymentRepository
	gateways    map[domain.GatewayType]domain.PaymentGateway
	lockSvc     *lock.RedisLock
	logger      *slog.Logger
}

// NewCallbackHandler 处理 HTTP 或 gRPC 请求。
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

// HandlePaymentCallback 处理支付结果异步回调
func (s *CallbackHandler) HandlePaymentCallback(ctx context.Context, paymentNo string, success bool, transactionID, thirdPartyNo string, callbackData map[string]string) error {
	s.logger.InfoContext(ctx, "handling payment callback", "payment_no", paymentNo, "success", success)

	// 1. 分布式锁，防止并发回调处理
	lockKey := fmt.Sprintf("lock:payment:callback:%s", paymentNo)
	token, err := s.lockSvc.Lock(ctx, lockKey, 10*time.Second)
	if err != nil {
		return fmt.Errorf("failed to acquire lock: %w", err)
	}
	defer s.lockSvc.Unlock(ctx, lockKey, token)

	// 2. 获取支付记录
	payment, err := s.paymentRepo.FindByPaymentNo(ctx, paymentNo)
	if err != nil || payment == nil {
		s.logger.ErrorContext(ctx, "payment not found", "payment_no", paymentNo)
		return fmt.Errorf("payment not found")
	}

	// 3. 幂等检查：如果已经是终态，则直接返回成功
	if payment.Status == domain.PaymentSuccess || payment.Status == domain.PaymentFailed {
		s.logger.InfoContext(ctx, "payment already processed", "payment_no", paymentNo, "status", payment.Status)
		return nil
	}

	// 4. 验证签名
	gateway, ok := s.gateways[payment.GatewayType]
	if ok {
		valid, err := gateway.VerifyCallback(ctx, callbackData)
		if err != nil || !valid {
			s.logger.WarnContext(ctx, "invalid callback signature", "payment_no", paymentNo)
			return fmt.Errorf("invalid signature")
		}
	}

	// 5. 领域层处理状态变更
	if err := payment.Process(success, transactionID, thirdPartyNo); err != nil {
		s.logger.ErrorContext(ctx, "failed to process state transition", "error", err)
		return err
	}

	// 6. 持久化
	if err := s.paymentRepo.Update(ctx, payment); err != nil {
		return err
	}

	s.logger.InfoContext(ctx, "payment callback processed successfully", "payment_no", paymentNo)
	return nil
}
