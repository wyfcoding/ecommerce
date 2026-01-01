package application

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/wyfcoding/ecommerce/internal/financialsettlement/domain"
)

// SettlementManager 处理财务结算模块的写操作和业务逻辑。
type SettlementManager struct {
	repo   domain.SettlementRepository
	logger *slog.Logger
}

// NewSettlementManager 创建并返回一个新的 SettlementManager 实例。
func NewSettlementManager(repo domain.SettlementRepository, logger *slog.Logger) *SettlementManager {
	return &SettlementManager{
		repo:   repo,
		logger: logger,
	}
}

// CreateSettlement 创建结算单。
func (m *SettlementManager) CreateSettlement(ctx context.Context, sellerID uint64, period string, startDate, endDate time.Time) (*domain.Settlement, error) {
	settlement := &domain.Settlement{
		SellerID:         sellerID,
		Period:           period,
		StartDate:        startDate,
		EndDate:          endDate,
		TotalSalesAmount: 100000,
		CommissionAmount: 5000,
		RebateAmount:     1000,
		OtherFees:        500,
		FinalAmount:      95500,
		Status:           domain.SettlementStatusPending,
	}

	if err := m.repo.SaveSettlement(ctx, settlement); err != nil {
		m.logger.Error("failed to create settlement", "error", err, "seller_id", sellerID)
		return nil, err
	}
	return settlement, nil
}

// ApproveSettlement 审核批准结算单。
func (m *SettlementManager) ApproveSettlement(ctx context.Context, id uint64, approvedBy string) error {
	settlement, err := m.repo.GetSettlement(ctx, id)
	if err != nil {
		return err
	}

	settlement.Status = domain.SettlementStatusApproved
	settlement.ApprovedBy = approvedBy
	now := time.Now()
	settlement.ApprovedAt = &now

	return m.repo.SaveSettlement(ctx, settlement)
}

// RejectSettlement 审核拒绝结算单。
func (m *SettlementManager) RejectSettlement(ctx context.Context, id uint64, reason string) error {
	settlement, err := m.repo.GetSettlement(ctx, id)
	if err != nil {
		return err
	}

	settlement.Status = domain.SettlementStatusRejected
	settlement.RejectionReason = reason

	return m.repo.SaveSettlement(ctx, settlement)
}

// ProcessPayment 处理结算代支付。
func (m *SettlementManager) ProcessPayment(ctx context.Context, settlementID uint64) (*domain.SettlementPayment, error) {
	settlement, err := m.repo.GetSettlement(ctx, settlementID)
	if err != nil {
		return nil, err
	}

	if settlement.Status != domain.SettlementStatusApproved {
		return nil, fmt.Errorf("settlement not approved")
	}

	payment := &domain.SettlementPayment{
		SettlementID:  settlementID,
		SellerID:      settlement.SellerID,
		Amount:        settlement.FinalAmount,
		Status:        domain.PaymentStatusProcessing,
		TransactionID: "TXN-" + time.Now().Format("20060102150405"),
	}

	if err := m.repo.SaveSettlementPayment(ctx, payment); err != nil {
		return nil, err
	}

	// 模拟支付完成
	payment.Status = domain.PaymentStatusCompleted
	now := time.Now()
	payment.CompletedAt = &now
	if err := m.repo.SaveSettlementPayment(ctx, payment); err != nil {
		m.logger.ErrorContext(ctx, "failed to update settlement payment status", "payment_id", payment.ID, "error", err)
	}

	// 更新结算单状态
	settlement.Status = domain.SettlementStatusCompleted
	if err := m.repo.SaveSettlement(ctx, settlement); err != nil {
		m.logger.ErrorContext(ctx, "failed to update settlement status", "settlement_id", settlement.ID, "error", err)
	}

	return payment, nil
}
