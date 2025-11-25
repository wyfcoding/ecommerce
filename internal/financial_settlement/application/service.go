package application

import (
	"context"
	"github.com/wyfcoding/ecommerce/internal/financial_settlement/domain/entity"
	"github.com/wyfcoding/ecommerce/internal/financial_settlement/domain/repository"
	"time"

	"log/slog"
)

type FinancialSettlementService struct {
	repo   repository.SettlementRepository
	logger *slog.Logger
}

func NewFinancialSettlementService(repo repository.SettlementRepository, logger *slog.Logger) *FinancialSettlementService {
	return &FinancialSettlementService{
		repo:   repo,
		logger: logger,
	}
}

// CreateSettlement 创建结算单
func (s *FinancialSettlementService) CreateSettlement(ctx context.Context, sellerID uint64, period string, startDate, endDate time.Time) (*entity.Settlement, error) {
	// In a real scenario, we would fetch orders and calculate amounts here
	// For now, we create a placeholder settlement
	settlement := &entity.Settlement{
		SellerID:         sellerID,
		Period:           period,
		StartDate:        startDate,
		EndDate:          endDate,
		TotalSalesAmount: 100000, // Mock
		CommissionAmount: 5000,   // Mock
		RebateAmount:     1000,   // Mock
		OtherFees:        500,    // Mock
		FinalAmount:      95500,  // Mock
		Status:           entity.SettlementStatusPending,
	}

	if err := s.repo.SaveSettlement(ctx, settlement); err != nil {
		s.logger.Error("failed to save settlement", "error", err)
		return nil, err
	}

	return settlement, nil
}

// ApproveSettlement 审核结算单
func (s *FinancialSettlementService) ApproveSettlement(ctx context.Context, id uint64, approvedBy string) error {
	settlement, err := s.repo.GetSettlement(ctx, id)
	if err != nil {
		return err
	}

	settlement.Status = entity.SettlementStatusApproved
	settlement.ApprovedBy = approvedBy
	now := time.Now()
	settlement.ApprovedAt = &now

	return s.repo.SaveSettlement(ctx, settlement)
}

// RejectSettlement 拒绝结算单
func (s *FinancialSettlementService) RejectSettlement(ctx context.Context, id uint64, reason string) error {
	settlement, err := s.repo.GetSettlement(ctx, id)
	if err != nil {
		return err
	}

	settlement.Status = entity.SettlementStatusRejected
	settlement.RejectionReason = reason

	return s.repo.SaveSettlement(ctx, settlement)
}

// GetSettlement 获取结算单详情
func (s *FinancialSettlementService) GetSettlement(ctx context.Context, id uint64) (*entity.Settlement, error) {
	return s.repo.GetSettlement(ctx, id)
}

// ListSettlements 获取结算单列表
func (s *FinancialSettlementService) ListSettlements(ctx context.Context, sellerID uint64, page, pageSize int) ([]*entity.Settlement, int64, error) {
	offset := (page - 1) * pageSize
	return s.repo.ListSellerSettlements(ctx, sellerID, offset, pageSize)
}

// ProcessPayment 处理支付
func (s *FinancialSettlementService) ProcessPayment(ctx context.Context, settlementID uint64) (*entity.SettlementPayment, error) {
	settlement, err := s.repo.GetSettlement(ctx, settlementID)
	if err != nil {
		return nil, err
	}

	if settlement.Status != entity.SettlementStatusApproved {
		return nil, ProcessError{Message: "Settlement not approved"}
	}

	payment := &entity.SettlementPayment{
		SettlementID:  settlementID,
		SellerID:      settlement.SellerID,
		Amount:        settlement.FinalAmount,
		Status:        entity.PaymentStatusProcessing,
		TransactionID: "TXN-" + time.Now().Format("20060102150405"),
	}

	if err := s.repo.SaveSettlementPayment(ctx, payment); err != nil {
		return nil, err
	}

	// Simulate payment completion
	payment.Status = entity.PaymentStatusCompleted
	now := time.Now()
	payment.CompletedAt = &now
	s.repo.SaveSettlementPayment(ctx, payment)

	settlement.Status = entity.SettlementStatusCompleted
	s.repo.SaveSettlement(ctx, settlement)

	return payment, nil
}

type ProcessError struct {
	Message string
}

func (e ProcessError) Error() string {
	return e.Message
}
