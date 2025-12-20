package application

import (
	"context"

	"github.com/wyfcoding/ecommerce/internal/dynamic_pricing/domain"
)

// DynamicPricingQuery handles read operations for dynamic pricing.
type DynamicPricingQuery struct {
	repo domain.PricingRepository
}

// NewDynamicPricingQuery creates a new DynamicPricingQuery instance.
func NewDynamicPricingQuery(repo domain.PricingRepository) *DynamicPricingQuery {
	return &DynamicPricingQuery{
		repo: repo,
	}
}

// GetLatestPrice 获取指定SKU的最新动态价格。
func (q *DynamicPricingQuery) GetLatestPrice(ctx context.Context, skuID uint64) (*domain.DynamicPrice, error) {
	return q.repo.GetLatestDynamicPrice(ctx, skuID)
}

// ListStrategies 获取定价策略列表。
func (q *DynamicPricingQuery) ListStrategies(ctx context.Context, page, pageSize int) ([]*domain.PricingStrategy, int64, error) {
	offset := (page - 1) * pageSize
	return q.repo.ListPricingStrategies(ctx, offset, pageSize)
}
