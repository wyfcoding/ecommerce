package application

import (
	"context"

	"github.com/wyfcoding/ecommerce/internal/settlement/domain"
)

// SettlementQuery 处理所有结算相关的查询操作（Queries）。
type SettlementQuery struct {
	repo domain.SettlementRepository
}

// NewSettlementQuery 构造函数。
func NewSettlementQuery(repo domain.SettlementRepository) *SettlementQuery {
	return &SettlementQuery{repo: repo}
}

// GetMerchantAccount 获取商户账户信息。
func (q *SettlementQuery) GetMerchantAccount(ctx context.Context, merchantID uint64) (*domain.MerchantAccount, error) {
	return q.repo.GetMerchantAccount(ctx, merchantID)
}

// ListSettlements 获取结算单列表。
func (q *SettlementQuery) ListSettlements(ctx context.Context, merchantID uint64, status *int, page, pageSize int) ([]*domain.Settlement, int64, error) {
	offset := (page - 1) * pageSize
	var st *domain.SettlementStatus
	if status != nil {
		s := domain.SettlementStatus(*status)
		st = &s
	}
	return q.repo.ListSettlements(ctx, merchantID, st, offset, pageSize)
}

// GetSettlement 获取单本结算单。
func (q *SettlementQuery) GetSettlement(ctx context.Context, id uint64) (*domain.Settlement, error) {
	return q.repo.GetSettlement(ctx, id)
}
