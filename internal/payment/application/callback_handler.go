// Package application 提供了支付模块的业务逻辑处理。
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

// CallbackHandler 专门负责处理来自外部支付网关（如支付宝、微信）的异步通知回调。
type CallbackHandler struct {
	paymentRepo domain.PaymentRepository                     // 支付数据仓储
	gateways    map[domain.GatewayType]domain.PaymentGateway // 网关适配器集合
	lockSvc     *lock.RedisLock                              // 分布式锁，用于并发处理保护
	outboxMgr   *outbox.Manager                              // 事务消息管理器
	logger      *slog.Logger                                 // 日志记录器
}

// NewCallbackHandler 创建并返回一个新的支付回调处理器实例。
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

// HandlePaymentCallback 执行支付结果回调的核心处理逻辑。
// 架构设计：分布式锁 -> 事务 -> 状态机流转 -> 发送事务消息。
func (s *CallbackHandler) HandlePaymentCallback(ctx context.Context, userID uint64, paymentNo string, success bool, transactionID, thirdPartyNo string, callbackData map[string]string) error {
	s.logger.InfoContext(ctx, "processing external payment callback", "payment_no", paymentNo, "user_id", userID, "success", success)

	// 1. 防重处理：利用分布式锁避免同一支付单的回调被多个节点重复执行
	lockKey := fmt.Sprintf("lock:payment:callback:%s", paymentNo)
	token, err := s.lockSvc.Lock(ctx, lockKey, 10*time.Second)
	if err != nil {
		s.logger.WarnContext(ctx, "failed to acquire callback lock, might be processing", "key", lockKey)
		return err
	}
	defer s.lockSvc.Unlock(ctx, lockKey, token)

	// 2. 事务性业务处理
	err = s.paymentRepo.Transaction(ctx, userID, func(tx any) error {
		txRepo := s.paymentRepo.WithTx(tx)
		payment, err := txRepo.FindByPaymentNo(ctx, userID, paymentNo)
		if err != nil || payment == nil {
			return fmt.Errorf("payment record not found: %s", paymentNo)
		}

		// 2.1 幂等性：若订单已是终态，则无需再次处理
		if payment.Status == domain.PaymentSuccess {
			s.logger.InfoContext(ctx, "callback ignored: payment already in terminal status", "payment_no", paymentNo)
			return nil
		}

		// 2.2 状态驱动逻辑：调用领域模型的 FSM 状态机进行合规流转
		if !success {
			return payment.Trigger(ctx, "CANCEL", "external gateway reported failure")
		}

		if err := payment.Trigger(ctx, "PAY_DIRECT", "confirmed via asynchronous callback"); err != nil {
			return fmt.Errorf("fsm transition failed: %w", err)
		}

		// 2.3 更新核心流水字段
		payment.TransactionID = transactionID
		payment.ThirdPartyNo = thirdPartyNo
		now := time.Now()
		payment.PaidAt = &now

		if err := txRepo.Update(ctx, payment); err != nil {
			return fmt.Errorf("failed to update payment record: %w", err)
		}

		// 2.4 发布支付成功事件至消息总线 (Outbox 模式保证 100% 投递)
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

	if err == nil {
		s.logger.InfoContext(ctx, "payment callback processed successfully", "payment_no", paymentNo)
	}
	return err
}
