package application

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/wyfcoding/ecommerce/internal/payment/domain"
	"github.com/wyfcoding/pkg/idgen"
	"github.com/wyfcoding/pkg/messagequeue/outbox"
	"github.com/wyfcoding/pkg/utils/ctxutil"
	"gorm.io/gorm"
)

type PaymentProcessor struct {
	paymentRepo domain.PaymentRepository
	channelRepo domain.ChannelRepository
	riskService domain.RiskService
	idGenerator idgen.Generator
	gateways    map[domain.GatewayType]domain.PaymentGateway
	outboxMgr   *outbox.Manager
	logger      *slog.Logger
}

func NewPaymentProcessor(
	paymentRepo domain.PaymentRepository,
	channelRepo domain.ChannelRepository,
	riskService domain.RiskService,
	idGenerator idgen.Generator,
	gateways map[domain.GatewayType]domain.PaymentGateway,
	outboxMgr *outbox.Manager,
	logger *slog.Logger,
) *PaymentProcessor {
	return &PaymentProcessor{
		paymentRepo: paymentRepo,
		channelRepo: channelRepo,
		riskService: riskService,
		idGenerator: idGenerator,
		gateways:    gateways,
		outboxMgr:   outboxMgr,
		logger:      logger,
	}
}

// InitiatePayment 这里的逻辑改为支持 PreAuth
func (s *PaymentProcessor) InitiatePayment(ctx context.Context, orderID uint64, userID uint64, amount int64, paymentMethodStr string) (*domain.Payment, *domain.PaymentGatewayResponse, error) {
	// 1. 获取网关
	gatewayType, _ := s.routeChannel(ctx, paymentMethodStr)
	gateway, ok := s.gateways[gatewayType]
	if !ok {
		return nil, nil, fmt.Errorf("unsupported gateway: %s", gatewayType)
	}

	// 2. 风控检查
	riskCtx := &domain.RiskContext{
		UserID:        userID,
		Amount:        amount,
		PaymentMethod: paymentMethodStr,
		IP:            ctxutil.GetIP(ctx),
		OrderID:       orderID,
	}
	riskResult, err := s.riskService.CheckPrePayment(ctx, riskCtx)
	if err != nil {
		s.logger.ErrorContext(ctx, "risk check failed", "error", err)
		return nil, nil, fmt.Errorf("risk check failed: %w", err)
	}

	if riskResult.Action == domain.RiskActionBlock {
		return nil, nil, fmt.Errorf("payment blocked by risk engine: %s", riskResult.Reason)
	}

	// 3. 创建或获取支付单
	payment, err := s.paymentRepo.FindByOrderID(ctx, userID, orderID)
	if err != nil {
		return nil, nil, err
	}
	if payment == nil {
		payment = domain.NewPayment(orderID, fmt.Sprintf("ORD%d", orderID), userID, amount, paymentMethodStr, gatewayType)
	}

	// 4. 执行 PreAuth
	gatewayReq := &domain.PaymentGatewayRequest{
		OrderID:     payment.PaymentNo,
		Amount:      payment.Amount,
		Currency:    "CNY",
		Description: "Order " + payment.OrderNo,
	}
	resp, err := gateway.PreAuth(ctx, gatewayReq)
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

	// 7. 保存
	if payment.ID == 0 {
		err = s.paymentRepo.Save(ctx, payment)
	} else {
		err = s.paymentRepo.Update(ctx, payment)
	}
	if err != nil {
		return nil, nil, err
	}

	return payment, resp, nil
}

// CapturePayment 捕获支付（确认支付）
func (s *PaymentProcessor) CapturePayment(ctx context.Context, userID uint64, paymentNo string, amount int64) error {
	return s.paymentRepo.Transaction(ctx, userID, func(tx any) error {
		txRepo := s.paymentRepo.WithTx(tx)
		payment, err := txRepo.FindByPaymentNo(ctx, userID, paymentNo)
		if err != nil || payment == nil {
			return fmt.Errorf("payment not found")
		}

		gateway, ok := s.gateways[payment.GatewayType]
		if !ok {
			return fmt.Errorf("gateway not found")
		}

		// 1. 调用网关 Capture
		_, err = gateway.Capture(ctx, payment.TransactionID, amount)
		if err != nil {
			return err
		}

		// 2. 状态变更
		if err := payment.Trigger(ctx, "CAPTURE", fmt.Sprintf("Captured amount: %d", amount)); err != nil {
			return err
		}
		payment.CapturedAmount = amount
		now := time.Now()
		payment.PaidAt = &now

		if err := txRepo.Update(ctx, payment); err != nil {
			return err
		}

		// 3. 发送结算事件 (Internal Service Interaction)
		event := map[string]any{
			"payment_no": payment.PaymentNo,
			"order_no":   payment.OrderNo,
			"user_id":    payment.UserID,
			"amount":     payment.CapturedAmount,
			"timestamp":  time.Now().Unix(),
		}
		gormTx := tx.(*gorm.DB)
		return s.outboxMgr.PublishInTx(gormTx, "payment.captured", payment.PaymentNo, event)
	})
}

func (s *PaymentProcessor) routeChannel(_ context.Context, method string) (domain.GatewayType, *domain.ChannelConfig) {
	// 简化处理，使用正确的 GatewayType 常量
	switch method {
	case "alipay":
		return domain.GatewayTypeAlipay, nil
	case "stripe":
		return domain.GatewayTypeStripe, nil
	default:
		return domain.GatewayTypeMock, nil
	}
}
