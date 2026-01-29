package application

import (
	"context"

	"github.com/wyfcoding/ecommerce/internal/settlement/domain"
)

// SettlementQueryService 处理所有结算相关的查询操作 (Queries)。
type SettlementQueryService struct {
	repo domain.SettlementRepository
}

// NewSettlementQueryService 构造函数。
func NewSettlementQueryService(repo domain.SettlementRepository) *SettlementQueryService {
	return &SettlementQueryService{repo: repo}
}

func (s *SettlementQueryService) GetSettlement(ctx context.Context, id uint64) (*domain.Settlement, error) {
	return s.repo.GetSettlement(ctx, id)
}

func (s *SettlementQueryService) GetSettlementByNo(ctx context.Context, no string) (*domain.Settlement, error) {
	return s.repo.GetSettlementByNo(ctx, no)
}

func (s *SettlementQueryService) ListSettlements(ctx context.Context, merchantID uint64, status *domain.SettlementStatus, page, pageSize int) ([]*domain.Settlement, int64, error) {
	offset := (page - 1) * pageSize
	return s.repo.ListSettlements(ctx, merchantID, status, offset, pageSize)
}

func (s *SettlementQueryService) ListSettlementDetails(ctx context.Context, settlementID uint64) ([]*domain.SettlementDetail, error) {
	return s.repo.ListSettlementDetails(ctx, settlementID)
}

func (s *SettlementQueryService) GetMerchantAccount(ctx context.Context, merchantID uint64) (*domain.MerchantAccount, error) {
	return s.repo.GetMerchantAccount(ctx, merchantID)
}
