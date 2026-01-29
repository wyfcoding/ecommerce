package application

import (
	"context"

	"github.com/wyfcoding/ecommerce/internal/dynamicpricing/domain"
)

// DynamicPricingQueryService 处理动态定价的读操作。
type DynamicPricingQueryService struct {
	repo domain.PricingRepository
}

// NewDynamicPricingQueryService 创建并返回一个新的 DynamicPricingQueryService 实例。
func NewDynamicPricingQueryService(repo domain.PricingRepository) *DynamicPricingQueryService {
	return &DynamicPricingQueryService{
		repo: repo,
	}
}

// GetLatestPrice 获取指定SKU的最新动态价格。
func (q *DynamicPricingQueryService) GetLatestPrice(ctx context.Context, skuID uint64) (*domain.DynamicPrice, error) {
	return q.repo.GetLatestDynamicPrice(ctx, skuID)
}

// ListStrategies 获取定价策略列表。
func (q *DynamicPricingQueryService) ListStrategies(ctx context.Context, page, pageSize int) ([]*domain.PricingStrategy, int64, error) {
	offset := (page - 1) * pageSize
	return q.repo.ListPricingStrategies(ctx, offset, pageSize)
}
