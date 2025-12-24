package application

import (
	"context"
	"time"

	"github.com/wyfcoding/ecommerce/internal/financialsettlement/domain"
)

// SettlementQuery 处理财务结算模块的查询操作。
type SettlementQuery struct {
	repo domain.SettlementRepository
}

// NewSettlementQuery 创建并返回一个新的 SettlementQuery 实例。
func NewSettlementQuery(repo domain.SettlementRepository) *SettlementQuery {
	return &SettlementQuery{repo: repo}
}

// GetSettlement 根据ID获取结算单详情。
func (q *SettlementQuery) GetSettlement(ctx context.Context, id uint64) (*domain.Settlement, error) {
	return q.repo.GetSettlement(ctx, id)
}

// ListSellerSettlements 获取卖家的结算单列表。
func (q *SettlementQuery) ListSellerSettlements(ctx context.Context, sellerID uint64, offset, limit int) ([]*domain.Settlement, int64, error) {
	return q.repo.ListSellerSettlements(ctx, sellerID, offset, limit)
}

// GetSettlementsByPeriod 获取指定周期的所有结算单。
func (q *SettlementQuery) GetSettlementsByPeriod(ctx context.Context, startDate, endDate time.Time) ([]*domain.Settlement, error) {
	return q.repo.GetSettlementsByPeriod(ctx, startDate, endDate)
}

// GetSettlementOrders 获取结算单明细。
func (q *SettlementQuery) GetSettlementOrders(ctx context.Context, settlementID uint64) ([]*domain.SettlementOrder, error) {
	return q.repo.GetSettlementOrders(ctx, settlementID)
}

// GetSettlementPayment 获取结算支付详情。
func (q *SettlementQuery) GetSettlementPayment(ctx context.Context, settlementID uint64) (*domain.SettlementPayment, error) {
	return q.repo.GetSettlementPaymentBySettlementID(ctx, settlementID)
}

// GetStatistics 获取结算统计。
func (q *SettlementQuery) GetStatistics(ctx context.Context, startDate, endDate time.Time) (*domain.SettlementStatistics, error) {
	return q.repo.GetSettlementStatistics(ctx, startDate, endDate)
}
