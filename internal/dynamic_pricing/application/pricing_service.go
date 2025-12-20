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

// NewDynamicPricingService creates a new DynamicPricingService facade.
func NewDynamicPricingService(manager *DynamicPricingManager, query *DynamicPricingQuery) *DynamicPricingService {
	return &DynamicPricingService{
		manager: manager,
		query:   query,
	}
}

// --- 写操作（委托给 Manager）---

func (s *DynamicPricingService) CalculatePrice(ctx context.Context, req *domain.PricingRequest) (*domain.DynamicPrice, error) {
	return s.manager.CalculatePrice(ctx, req)
}

func (s *DynamicPricingService) SaveStrategy(ctx context.Context, strategy *domain.PricingStrategy) error {
	return s.manager.SaveStrategy(ctx, strategy)
}

// --- 读操作（委托给 Query）---

func (s *DynamicPricingService) GetLatestPrice(ctx context.Context, skuID uint64) (*domain.DynamicPrice, error) {
	return s.query.GetLatestPrice(ctx, skuID)
}

func (s *DynamicPricingService) ListStrategies(ctx context.Context, page, pageSize int) ([]*domain.PricingStrategy, int64, error) {
	return s.query.ListStrategies(ctx, page, pageSize)
}
