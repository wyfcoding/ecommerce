package application

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	pb "github.com/wyfcoding/ecommerce/goapi/payment/v1"
	"github.com/wyfcoding/ecommerce/internal/payment/domain"
	"github.com/wyfcoding/pkg/contextx"
	"github.com/wyfcoding/pkg/idgen"
	"github.com/wyfcoding/pkg/lock"
	"github.com/wyfcoding/pkg/messagequeue"
)

type PaymentCommandService struct {
	paymentRepo domain.PaymentRepository
	refundRepo  domain.RefundRepository
	channelRepo domain.ChannelRepository
	eventStore  domain.EventStore
	routing     *RoutingEngine
	riskService domain.RiskService
	idGenerator idgen.Generator
	gateways    map[domain.GatewayType]domain.PaymentGateway
	publisher   messagequeue.EventPublisher
	lockSvc     *lock.RedisLock
	logger      *slog.Logger
}

func NewPaymentCommandService(
	paymentRepo domain.PaymentRepository,
	refundRepo domain.RefundRepository,
	channelRepo domain.ChannelRepository,
	eventStore domain.EventStore,
	riskService domain.RiskService,
	idGenerator idgen.Generator,
	gateways map[domain.GatewayType]domain.PaymentGateway,
	publisher messagequeue.EventPublisher,
	lockSvc *lock.RedisLock,
	logger *slog.Logger,
) *PaymentCommandService {
	return &PaymentCommandService{
		paymentRepo: paymentRepo,
		refundRepo:  refundRepo,
		channelRepo: channelRepo,
		eventStore:  eventStore,
		routing:     NewRoutingEngine(channelRepo),
		riskService: riskService,
		idGenerator: idGenerator,
		gateways:    gateways,
		publisher:   publisher,
		lockSvc:     lockSvc,
		logger:      logger,
	}
}

// InitiatePayment 顶级架构：支持智能路由与自动化分账
func (s *PaymentCommandService) InitiatePayment(ctx context.Context, cmd *InitiatePaymentCommand) (*domain.Payment, *domain.PaymentGatewayResponse, error) {
	// 1. 智能路由决策 (Adyen Standard)
	gatewayType, chCfg := s.routing.SelectBestChannel(ctx, cmd.Amount, cmd.PaymentMethod)
	gateway, ok := s.gateways[gatewayType]
	if !ok {
		return nil, nil, fmt.Errorf("unsupported gateway path: %s", gatewayType)
	}

	// 2. 深度风控检查 (Ant Group Level)
	riskCtx := &domain.RiskContext{
		UserID: cmd.UserID, Amount: cmd.Amount, PaymentMethod: cmd.PaymentMethod,
		IP: cmd.ClientIP, OrderID: cmd.OrderID, DeviceID: cmd.DeviceID,
	}
	if riskCtx.IP == "" {
		riskCtx.IP = contextx.GetIP(ctx)
	}
	if riskCtx.DeviceID == "" {
		riskCtx.DeviceID = contextx.GetUserAgent(ctx)
	}

	riskResult, err := s.riskService.CheckPrePayment(ctx, riskCtx)
	if err != nil {
		s.logger.ErrorContext(ctx, "risk check failed", "error", err)
		return nil, nil, fmt.Errorf("risk check failed: %w", err)
	}

	if riskResult.Action == domain.RiskActionBlock {
		return nil, nil, fmt.Errorf("high risk blocked: %s", riskResult.Reason)
	}

	payment, err := s.paymentRepo.FindByOrderID(ctx, cmd.UserID, cmd.OrderID)
	if err != nil {
		return nil, nil, err
	}
	if payment == nil {
		payment = domain.NewPayment(cmd.OrderID, fmt.Sprintf("ORD%d", cmd.OrderID), cmd.UserID, cmd.Amount, cmd.PaymentMethod, gatewayType, s.idGenerator)
	}

	// 4. 执行网关 PreAuth 并记录指标
	start := time.Now()
	gatewayReq := &domain.PaymentGatewayRequest{
		OrderID: payment.PaymentNo, UserID: cmd.UserID, Amount: payment.Amount, Currency: "CNY",
	}
	resp, err := gateway.PreAuth(ctx, gatewayReq)
	s.routing.RecordResult(chCfg.Code, err == nil, time.Since(start))

	if err != nil {
		return nil, nil, err
	}

	// 5. 记录风控交易 (用于频控)
	if err := s.riskService.RecordTransaction(ctx, riskCtx); err != nil {
		s.logger.WarnContext(ctx, "failed to record risk transaction", "error", err)
	}

	// 6. 更新领域模型状态
	if err := payment.Trigger(ctx, "AUTH", "Pre-authorization successful"); err != nil {
		return nil, nil, err
	}
	payment.TransactionID = resp.TransactionID

	// 7. 保存聚合与事件
	if err := s.saveAggregate(ctx, payment); err != nil {
		return nil, nil, err
	}

	s.logger.InfoContext(ctx, "payment initiated successfully", "payment_no", payment.PaymentNo, "transaction_id", resp.TransactionID)
	return payment, resp, nil
}

// saveAggregate 内部辅助方法，保存聚合状态并持久化未提交的事件。
func (s *PaymentCommandService) saveAggregate(ctx context.Context, p *domain.Payment) error {
	events := p.GetUncommittedEvents()
	if len(events) == 0 {
		return nil
	}

	return s.paymentRepo.Transaction(ctx, p.UserID, func(tx any) error {
		txRepo := s.paymentRepo.WithTx(tx)

		var err error
		if p.ID == 0 {
			err = txRepo.Save(ctx, p)
		} else {
			err = txRepo.Update(ctx, p)
		}
		if err != nil {
			return err
		}

		// 保存事件到事件存储。
		if err := s.eventStore.Save(ctx, events); err != nil {
			return err
		}

		// 发布事件。
		for _, e := range events {
			topic := s.getTopicForEvent(e)
			if topic != "" {
				if err := s.publisher.PublishInTx(ctx, tx, topic, p.PaymentNo, e); err != nil {
					return err
				}
			}
		}

		p.MarkCommitted()
		return nil
	})
}

func (s *PaymentCommandService) getTopicForEvent(event any) string {
	switch event.(type) {
	case *domain.PaymentInitiatedEvent:
		return "payment.initiated"
	case *domain.PaymentAuthorizedEvent:
		return "payment.authorized"
	case *domain.PaymentCapturedEvent:
		return "payment.captured"
	case *domain.PaymentPaidEvent:
		return "payment.paid"
	case *domain.RefundFinishedEvent:
		return "payment.refunded"
	case *domain.PaymentClosedEvent:
		return "payment.closed"
	}
	return ""
}

// CapturePayment 对标金融级账本一致性
func (s *PaymentCommandService) CapturePayment(ctx context.Context, cmd *CapturePaymentCommand) error {
	err := s.paymentRepo.Transaction(ctx, cmd.UserID, func(tx any) error {
		txRepo := s.paymentRepo.WithTx(tx)
		payment, err := txRepo.FindByPaymentNo(ctx, cmd.UserID, cmd.PaymentNo)
		if err != nil || payment == nil {
			return fmt.Errorf("payment not found")
		}

		gateway, ok := s.gateways[payment.GatewayType]
		if !ok {
			return fmt.Errorf("gateway not found")
		}

		// 1. 网关 Capture
		_, err = gateway.Capture(ctx, payment.TransactionID, cmd.Amount)
		if err != nil {
			return err
		}

		// 2. 状态驱动变更 (FSM) - 事件将包含 PaidAt 等信息
		payment.CapturedAmount = cmd.Amount
		if err := payment.Trigger(ctx, "CAPTURE", "Real-time fund capture"); err != nil {
			return err
		}

		// 3. 更新分账状态
		for i := range payment.Splits {
			payment.Splits[i].Status = "SETTLED"
		}

		// 4. 保存聚合 (含事件发布)
		return s.saveAggregate(ctx, payment)
	})

	if err == nil {
		s.logger.InfoContext(ctx, "payment captured successfully", "payment_no", cmd.PaymentNo, "amount", cmd.Amount)
	}
	return err
}

// RequestRefund 处理退款请求
func (s *PaymentCommandService) RequestRefund(ctx context.Context, cmd *RefundPaymentCommand) (*domain.Refund, error) {
	payment, err := s.paymentRepo.FindByID(ctx, cmd.UserID, cmd.PaymentID)
	if err != nil || payment == nil {
		return nil, fmt.Errorf("payment not found")
	}

	// 1. 调用网关退款
	gateway, ok := s.gateways[payment.GatewayType]
	if !ok {
		return nil, fmt.Errorf("gateway not found: %s", payment.GatewayType)
	}

	if err := gateway.Refund(ctx, payment.TransactionID, cmd.Amount); err != nil {
		return nil, err
	}

	var refund *domain.Refund
	// 2. 事务处理
	err = s.paymentRepo.Transaction(ctx, cmd.UserID, func(tx any) error {
		txPaymentRepo := s.paymentRepo.WithTx(tx)
		txRefundRepo := s.refundRepo.WithTx(tx)

		// 重新加载支付单状态
		p, err := txPaymentRepo.FindByID(ctx, cmd.UserID, cmd.PaymentID)
		if err != nil {
			return err
		}

		// 状态机处理 (退款申请)
		if err := p.Trigger(ctx, "REFUND_REQ", cmd.Reason); err != nil {
			return err
		}

		// 创建退款单
		refund = &domain.Refund{
			RefundNo:     fmt.Sprintf("REF%d", s.idGenerator.Generate()),
			PaymentID:    uint64(p.ID),
			PaymentNo:    p.PaymentNo,
			OrderID:      p.OrderID,
			OrderNo:      p.OrderNo,
			UserID:       p.UserID,
			RefundAmount: cmd.Amount,
			Reason:       cmd.Reason,
			Status:       p.Status,
		}

		// 状态机处理 (退款完成)
		if err := p.Trigger(ctx, "REFUND_FINISH", "Refund completed"); err != nil {
			return err
		}

		refund.Status = p.Status
		now := time.Now()
		refund.RefundedAt = &now

		// 保存支付单和退款单
		if err := txRefundRepo.Save(ctx, refund); err != nil {
			return err
		}
		return s.saveAggregate(ctx, p)
	})
	if err != nil {
		return nil, err
	}

	return refund, nil
}

// HandlePaymentCallback 执行支付结果回调的核心处理逻辑。
// 架构设计：分布式锁 -> 事务 -> 状态机流转 -> 发送事务消息。
func (s *PaymentCommandService) HandlePaymentCallback(ctx context.Context, userID uint64, paymentNo string, success bool, transactionID, thirdPartyNo string, callbackData map[string]string) error {
	s.logger.InfoContext(ctx, "processing external payment callback", "payment_no", paymentNo, "user_id", userID, "success", success)

	// 1. 防重处理：利用分布式锁避免同一支付单的回调被多个节点重复执行
	lockKey := fmt.Sprintf("lock:payment:callback:%s", paymentNo)
	token, err := s.lockSvc.Lock(ctx, lockKey, 10*time.Second)
	if err != nil {
		s.logger.WarnContext(ctx, "failed to acquire callback lock, might be processing", "key", lockKey)
		return err
	}
	defer func() {
		if uErr := s.lockSvc.Unlock(ctx, lockKey, token); uErr != nil {
			s.logger.WarnContext(ctx, "failed to release callback lock", "key", lockKey, "error", uErr)
		}
	}()

	// 2. 事务性业务处理
	err = s.paymentRepo.Transaction(ctx, userID, func(tx any) error {
		txRepo := s.paymentRepo.WithTx(tx)
		payment, err := txRepo.FindByPaymentNo(ctx, userID, paymentNo)
		if err != nil || payment == nil {
			return fmt.Errorf("payment record not found: %s", paymentNo)
		}

		// 2.1 幂等性：若订单已是终态，则无需再次处理
		if payment.Status == pb.PaymentStatus_SUCCESS {
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

		// 2.4 保存聚合 (含事件发布)
		return s.saveAggregate(ctx, payment)
	})

	if err == nil {
		s.logger.InfoContext(ctx, "payment callback processed successfully", "payment_no", paymentNo)
	}
	return err
}

// SagaRefund Saga 正向: 执行退款 (原路退回)
func (s *PaymentCommandService) SagaRefund(ctx context.Context, barrier any, userID, orderID uint64, amount int64, reason string) (string, error) {
	var refundNo string
	err := s.paymentRepo.ExecWithBarrier(ctx, barrier, func(ctx context.Context) error {
		// 1. 查找支付单
		payment, err := s.paymentRepo.FindByOrderID(ctx, userID, orderID)
		if err != nil || payment == nil {
			return fmt.Errorf("payment not found for order %d", orderID)
		}

		// 2. 调用网关退款
		gateway, ok := s.gateways[payment.GatewayType]
		if !ok {
			return fmt.Errorf("gateway not found: %s", payment.GatewayType)
		}

		if err := gateway.Refund(ctx, payment.TransactionID, amount); err != nil {
			return fmt.Errorf("gateway refund failed: %w", err)
		}

		// 3. 记录内部退款流水 (状态：REFUNDED)
		refundNo = fmt.Sprintf("SAGA-REF-%d", s.idGenerator.Generate())
		refund := &domain.Refund{
			RefundNo:     refundNo,
			PaymentID:    uint64(payment.ID),
			PaymentNo:    payment.PaymentNo,
			OrderID:      orderID,
			OrderNo:      payment.OrderNo,
			UserID:       userID,
			RefundAmount: amount,
			Reason:       reason,
			Status:       pb.PaymentStatus_REFUNDED,
		}

		now := time.Now()
		refund.RefundedAt = &now

		return s.refundRepo.Save(ctx, refund)
	})
	return refundNo, err
}

// SagaCancelRefund Saga 补偿: 记录退款异常
func (s *PaymentCommandService) SagaCancelRefund(ctx context.Context, barrier any, userID, orderID uint64) error {
	return s.paymentRepo.ExecWithBarrier(ctx, barrier, func(ctx context.Context) error {
		s.logger.WarnContext(ctx, "SagaCancelRefund called! Manual intervention may be needed for order", "order_id", orderID)
		return nil
	})
}
