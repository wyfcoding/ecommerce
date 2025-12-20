package application

import (
	"context"

	"github.com/wyfcoding/ecommerce/internal/pricing/domain"
)

// PricingService acts as a facade for pricing operations.
type PricingService struct {
	manager *PricingManager
	query   *PricingQuery
}

// NewPricingService creates a new PricingService facade.
func NewPricingService(manager *PricingManager, query *PricingQuery) *PricingService {
	return &PricingService{
		manager: manager,
		query:   query,
	}
}

// --- Write Operations (Delegated to Manager) ---

func (s *PricingService) CreateRule(ctx context.Context, rule *domain.PricingRule) error {
	return s.manager.CreateRule(ctx, rule)
}

func (s *PricingService) RecordHistory(ctx context.Context, productID, skuID, price, oldPrice uint64, reason string) error {
	return s.manager.RecordHistory(ctx, productID, skuID, price, oldPrice, reason)
}

// --- Read Operations (Delegated to Query) ---

func (s *PricingService) CalculatePrice(ctx context.Context, productID, skuID uint64, demand, competition float64) (uint64, error) {
	return s.query.CalculatePrice(ctx, productID, skuID, demand, competition)
}

func (s *PricingService) ListRules(ctx context.Context, productID uint64, page, pageSize int) ([]*domain.PricingRule, int64, error) {
	return s.query.ListRules(ctx, productID, page, pageSize)
}

func (s *PricingService) ListHistory(ctx context.Context, productID, skuID uint64, page, pageSize int) ([]*domain.PriceHistory, int64, error) {
	return s.query.ListHistory(ctx, productID, skuID, page, pageSize)
}
