package application

import (
	"context"
	"time"

	"github.com/wyfcoding/ecommerce/internal/settlement/domain"
)

// SettlementService 是结算应用服务的门面。
type SettlementService struct {
	Command *SettlementCommandService
	Query   *SettlementQueryService
}

// NewSettlementService 创建结算服务门面实例。
func NewSettlementService(command *SettlementCommandService, query *SettlementQueryService) *SettlementService {
	return &SettlementService{
		Command: command,
		Query:   query,
	}
}

// --- Delegate Command Methods ---

func (s *SettlementService) CreateSettlement(ctx context.Context, merchantID uint64, cycle string, startDate, endDate time.Time) (*domain.Settlement, error) {
	return s.Command.CreateSettlement(ctx, merchantID, cycle, startDate, endDate)
}

func (s *SettlementService) RecordPaymentSuccess(ctx context.Context, orderID uint64, orderNo string, merchantID uint64, amount int64, fee int64) error {
	return s.Command.RecordPaymentSuccess(ctx, orderID, orderNo, merchantID, amount, fee)
}

func (s *SettlementService) AddOrderToSettlement(ctx context.Context, settlementID uint64, orderID uint64, orderNo string, amount uint64) error {
	return s.Command.AddOrderToSettlement(ctx, settlementID, orderID, orderNo, amount)
}

func (s *SettlementService) ProcessSettlement(ctx context.Context, id uint64) error {
	return s.Command.ProcessSettlement(ctx, id)
}

func (s *SettlementService) CompleteSettlement(ctx context.Context, id uint64) error {
	return s.Command.CompleteSettlement(ctx, id)
}

// --- Delegate Query Methods ---

func (s *SettlementService) GetSettlement(ctx context.Context, id uint64) (*domain.Settlement, error) {
	return s.Query.GetSettlement(ctx, id)
}

func (s *SettlementService) ListSettlements(ctx context.Context, merchantID uint64, status *domain.SettlementStatus, page, pageSize int) ([]*domain.Settlement, int64, error) {
	return s.Query.ListSettlements(ctx, merchantID, status, page, pageSize)
}

func (s *SettlementService) GetMerchantAccount(ctx context.Context, merchantID uint64) (*domain.MerchantAccount, error) {
	return s.Query.GetMerchantAccount(ctx, merchantID)
}
