package application

import (
	"context"

	"github.com/wyfcoding/ecommerce/internal/dynamic_pricing/domain"
)

// DynamicPricingService 作为动态定价操作的门面。
type DynamicPricingService struct {
	manager *DynamicPricingManager
	query   *DynamicPricingQuery
}

// NewDynamicPricingService 创建动态定价服务门面实例。
func NewDynamicPricingService(manager *DynamicPricingManager, query *DynamicPricingQuery) *DynamicPricingService {
	return &DynamicPricingService{
		manager: manager,
		query:   query,
	}
}

// --- 写操作（委托给 Manager）---

// CalculatePrice 核心算法：基于实时策略计算动态价格。
func (s *DynamicPricingService) CalculatePrice(ctx context.Context, req *domain.PricingRequest) (*domain.DynamicPrice, error) {
	return s.manager.CalculatePrice(ctx, req)
}

// SaveStrategy 保存或更新动态定价策略配置。
func (s *DynamicPricingService) SaveStrategy(ctx context.Context, strategy *domain.PricingStrategy) error {
	return s.manager.SaveStrategy(ctx, strategy)
}

// --- 读操作（委托给 Query）---

// GetLatestPrice 获取指定SKU的最新动态价格快照。
func (s *DynamicPricingService) GetLatestPrice(ctx context.Context, skuID uint64) (*domain.DynamicPrice, error) {
	return s.query.GetLatestPrice(ctx, skuID)
}

// ListStrategies 获取所有定价策略列表（分页）。
func (s *DynamicPricingService) ListStrategies(ctx context.Context, page, pageSize int) ([]*domain.PricingStrategy, int64, error) {
	return s.query.ListStrategies(ctx, page, pageSize)
}
