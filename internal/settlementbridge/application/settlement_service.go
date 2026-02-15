// Package application 资金结算桥接应用服务
package application

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/wyfcoding/ecommerce/internal/settlementbridge/domain"
	"github.com/wyfcoding/pkg/messagequeue"
)

// SettlementService 结算桥接服务
type SettlementService struct {
	repo          domain.SettlementRepository
	treasurySvc   domain.TreasuryService
	paymentSvc    domain.PaymentService
	eventConsumer messagequeue.EventConsumer
	logger        *slog.Logger
}

func NewSettlementService(
	repo domain.SettlementRepository,
	treasurySvc domain.TreasuryService,
	paymentSvc domain.PaymentService,
	eventConsumer messagequeue.EventConsumer,
	logger *slog.Logger,
) *SettlementService {
	return &SettlementService{
		repo:          repo,
		treasurySvc:   treasurySvc,
		paymentSvc:    paymentSvc,
		eventConsumer: eventConsumer,
		logger:        logger.With("module", "settlement_service"),
	}
}

// StartEventListening 启动事件监听
func (s *SettlementService) StartEventListening(ctx context.Context) error {
	// 监听电商支付成功事件
	return s.eventConsumer.Subscribe(ctx, "payment.succeeded", s.handlePaymentSucceeded)
}

// handlePaymentSucceeded 处理支付成功事件
func (s *SettlementService) handlePaymentSucceeded(ctx context.Context, event messagequeue.Event) error {
	paymentID, ok := event.Data["payment_id"].(string)
	if !ok {
		return fmt.Errorf("invalid payment_id in event")
	}

	// 获取支付详情
	payment, err := s.paymentSvc.GetPaymentDetails(ctx, paymentID)
	if err != nil {
		return fmt.Errorf("failed to get payment details: %w", err)
	}

	// 创建结算桥接记录
	settlement := domain.NewSettlementBridge(
		payment.PaymentID,
		payment.OrderID,
		payment.Amount,
		payment.Currency,
	)

	// 设置账户映射（这里简化处理，实际需要配置映射规则）
	settlement.FromAccount = "ECOMMERCE_CASH_POOL"
	settlement.ToAccount = "FINANCIAL_MAIN_POOL"

	if err := s.repo.Save(ctx, settlement); err != nil {
		return fmt.Errorf("failed to save settlement: %w", err)
	}

	// 立即尝试处理结算
	go s.processSettlement(context.Background(), settlement)

	return nil
}

// processSettlement 处理单笔结算
func (s *SettlementService) processSettlement(ctx context.Context, settlement *domain.SettlementBridge) {
	settlement.StartProcessing()
	_ = s.repo.Save(ctx, settlement)

	// 调用金融国库服务进行资金转移
	err := s.treasurySvc.DepositToTreasury(
		ctx,
		settlement.ToAccount,
		settlement.Amount,
		settlement.Currency,
		settlement.PaymentID,
	)

	if err != nil {
		s.logger.ErrorContext(ctx, "settlement failed",
			"payment_id", settlement.PaymentID,
			"error", err)
		settlement.Fail(err.Error())
	} else {
		s.logger.InfoContext(ctx, "settlement completed",
			"payment_id", settlement.PaymentID,
			"amount", settlement.Amount)
		settlement.Complete()
	}

	_ = s.repo.Save(ctx, settlement)
}

// ProcessPendingSettlements 处理待处理的结算记录
func (s *SettlementService) ProcessPendingSettlements(ctx context.Context) error {
	pending, err := s.repo.ListPending(ctx, 100)
	if err != nil {
		return err
	}

	for _, settlement := range pending {
		go s.processSettlement(ctx, settlement)
	}

	return nil
}

// GetSettlementStatus 查询结算状态
func (s *SettlementService) GetSettlementStatus(ctx context.Context, paymentID string) (*domain.SettlementBridge, error) {
	return s.repo.GetByPaymentID(ctx, paymentID)
}
