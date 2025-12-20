package application

import (
	"context"
	"errors"

	"github.com/wyfcoding/ecommerce/internal/pricing/domain"
)

// PricingQuery 处理读操作和计算。
type PricingQuery struct {
	repo domain.PricingRepository
}

// NewPricingQuery creates a new PricingQuery instance.
func NewPricingQuery(repo domain.PricingRepository) *PricingQuery {
	return &PricingQuery{
		repo: repo,
	}
}

// CalculatePrice 根据定价规则计算商品或SKU的价格。
func (q *PricingQuery) CalculatePrice(ctx context.Context, productID, skuID uint64, demand, competition float64) (uint64, error) {
	rule, err := q.repo.GetActiveRule(ctx, productID, skuID)
	if err != nil {
		return 0, err
	}
	if rule == nil {
		return 0, errors.New("no active pricing rule found")
	}

	price := rule.CalculatePrice(demand, competition)
	return price, nil
}

// ListRules 获取定价规则列表。
func (q *PricingQuery) ListRules(ctx context.Context, productID uint64, page, pageSize int) ([]*domain.PricingRule, int64, error) {
	offset := (page - 1) * pageSize
	return q.repo.ListRules(ctx, productID, offset, pageSize)
}

// ListHistory 获取价格历史记录列表。
func (q *PricingQuery) ListHistory(ctx context.Context, productID, skuID uint64, page, pageSize int) ([]*domain.PriceHistory, int64, error) {
	offset := (page - 1) * pageSize
	return q.repo.ListHistory(ctx, productID, skuID, offset, pageSize)
}
