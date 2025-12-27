package application

import (
	"context"
	"log/slog"
	"time"

	"github.com/wyfcoding/ecommerce/internal/settlement/domain"
)

// SettlementService 结算门面服务，整合 Manager 和 Query。
type SettlementService struct {
	manager *SettlementManager
	query   *SettlementQuery
}

// NewSettlementService 构造函数。
func NewSettlementService(
	repo domain.SettlementRepository,
	ledgerService *domain.LedgerService,
	logger *slog.Logger,
) *SettlementService {
	return &SettlementService{
		manager: NewSettlementManager(repo, ledgerService, logger),
		query:   NewSettlementQuery(repo),
	}
}

// --- Manager (Writes) ---

func (s *SettlementService) RecordPaymentSuccess(ctx context.Context, orderID uint64, orderNo string, merchantID uint64, amount int64, channelCost int64) error {
	return s.manager.RecordPaymentSuccess(ctx, orderID, orderNo, merchantID, amount, channelCost)
}

func (s *SettlementService) CreateSettlement(ctx context.Context, merchantID uint64, cycle string, startDate, endDate time.Time) (*domain.Settlement, error) {
	return s.manager.CreateSettlement(ctx, merchantID, cycle, startDate, endDate)
}

func (s *SettlementService) AddOrderToSettlement(ctx context.Context, settlementID uint64, orderID uint64, orderNo string, amount uint64) error {
	return s.manager.AddOrderToSettlement(ctx, settlementID, orderID, orderNo, amount)
}

func (s *SettlementService) ProcessSettlement(ctx context.Context, id uint64) error {
	return s.manager.ProcessSettlement(ctx, id)
}

func (s *SettlementService) CompleteSettlement(ctx context.Context, id uint64) error {
	return s.manager.CompleteSettlement(ctx, id)
}

// --- Query (Reads) ---

func (s *SettlementService) GetMerchantAccount(ctx context.Context, merchantID uint64) (*domain.MerchantAccount, error) {
	return s.query.GetMerchantAccount(ctx, merchantID)
}

func (s *SettlementService) ListSettlements(ctx context.Context, merchantID uint64, status *int, page, pageSize int) ([]*domain.Settlement, int64, error) {
	return s.query.ListSettlements(ctx, merchantID, status, page, pageSize)
}

func (s *SettlementService) GetSettlement(ctx context.Context, id uint64) (*domain.Settlement, error) {
	return s.query.GetSettlement(ctx, id)
}
