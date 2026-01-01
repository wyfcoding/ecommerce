package application

import (
	"context"

	"github.com/wyfcoding/ecommerce/internal/dynamicpricing/domain"
)

// DynamicPricingQuery 处理动态定价的读操作。
type DynamicPricingQuery struct {
	repo domain.PricingRepository
}

// NewDynamicPricingQuery 创建并返回一个新的 DynamicPricingQuery 实例。
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
