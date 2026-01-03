package application

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/wyfcoding/ecommerce/internal/payment/domain"
	"github.com/wyfcoding/pkg/lock"
	"github.com/wyfcoding/pkg/messagequeue/outbox"
	"gorm.io/gorm"
)

type CallbackHandler struct {
	paymentRepo domain.PaymentRepository
	gateways    map[domain.GatewayType]domain.PaymentGateway
	lockSvc     *lock.RedisLock
	outboxMgr   *outbox.Manager
	logger      *slog.Logger
}

func NewCallbackHandler(
	paymentRepo domain.PaymentRepository,
	gateways map[domain.GatewayType]domain.PaymentGateway,
	lockSvc *lock.RedisLock,
	outboxMgr *outbox.Manager,
	logger *slog.Logger,
) *CallbackHandler {
	return &CallbackHandler{
		paymentRepo: paymentRepo,
		gateways:    gateways,
		lockSvc:     lockSvc,
		outboxMgr:   outboxMgr,
		logger:      logger,
	}
}

// HandlePaymentCallback 处理支付回调逻辑
func (s *CallbackHandler) HandlePaymentCallback(ctx context.Context, userID uint64, paymentNo string, success bool, transactionID, thirdPartyNo string, callbackData map[string]string) error {
	s.logger.InfoContext(ctx, "handling payment callback", "payment_no", paymentNo, "user_id", userID, "success", success)

	// 1. 分布式锁保护，防止并发回调处理
	lockKey := fmt.Sprintf("lock:payment:callback:%s", paymentNo)
	token, err := s.lockSvc.Lock(ctx, lockKey, 10*time.Second)
	if err != nil {
		return err
	}
	defer s.lockSvc.Unlock(ctx, lockKey, token)

	// 2. 本地事务处理
	return s.paymentRepo.Transaction(ctx, userID, func(tx any) error {
		txRepo := s.paymentRepo.WithTx(tx)
		payment, err := txRepo.FindByPaymentNo(ctx, userID, paymentNo)
		if err != nil || payment == nil {
			return fmt.Errorf("payment not found")
		}

		// 2.1 幂等性检查：如果已处理，则直接返回
		if payment.Status == domain.PaymentSuccess {
			s.logger.InfoContext(ctx, "payment already success, skipping callback", "payment_no", paymentNo)
			return nil
		}

		// 2.2 状态机更新
		if !success {
			return payment.Trigger(ctx, "CANCEL", "Gateway reported failure")
		}

		if err := payment.Trigger(ctx, "PAY_DIRECT", "External callback confirm"); err != nil {
			return err
		}
		
		payment.TransactionID = transactionID
		payment.ThirdPartyNo = thirdPartyNo
		now := time.Now()
		payment.PaidAt = &now

		if err := txRepo.Update(ctx, payment); err != nil {
			return err
		}

		// 3. 发布可靠的支付成功事件
		// 此事件由订单服务订阅，用于自动改为“已支付”状态
		event := map[string]any{
			"payment_no": payment.PaymentNo,
			"order_no":   payment.OrderNo,
			"user_id":    payment.UserID,
			"amount":     payment.Amount,
			"paid_at":    now.Unix(),
		}
		gormTx := tx.(*gorm.DB)
		return s.outboxMgr.PublishInTx(ctx, gormTx, "payment.paid", payment.PaymentNo, event)
	})
}
