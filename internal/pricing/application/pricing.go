package application

import (
	"context"

	"github.com/wyfcoding/ecommerce/internal/pricing/domain"
)

// PricingService 作为定价操作的门面。
type PricingService struct {
	manager *PricingManager
	query   *PricingQuery
}

// NewPricingService 创建定价服务门面实例。
func NewPricingService(manager *PricingManager, query *PricingQuery) *PricingService {
	return &PricingService{
		manager: manager,
		query:   query,
	}
}

// --- 写操作（委托给 Manager）---

// CreateRule 创建一个新的定价规则。
func (s *PricingService) CreateRule(ctx context.Context, rule *domain.PricingRule) error {
	return s.manager.CreateRule(ctx, rule)
}

// RecordHistory 记录价格变更历史。
func (s *PricingService) RecordHistory(ctx context.Context, productID, skuID, price, oldPrice uint64, reason string) error {
	return s.manager.RecordHistory(ctx, productID, skuID, price, oldPrice, reason)
}

// --- 读操作（委托给 Query）---

// CalculatePrice 计算动态价格。
func (s *PricingService) CalculatePrice(ctx context.Context, productID, skuID uint64, demand, competition float64) (uint64, error) {
	return s.query.CalculatePrice(ctx, productID, skuID, demand, competition)
}

// ListRules 列出定价规则。
func (s *PricingService) ListRules(ctx context.Context, productID uint64, page, pageSize int) ([]*domain.PricingRule, int64, error) {
	return s.query.ListRules(ctx, productID, page, pageSize)
}

// ListHistory 获取价格变更历史列表。
func (s *PricingService) ListHistory(ctx context.Context, productID, skuID uint64, page, pageSize int) ([]*domain.PriceHistory, int64, error) {
	return s.query.ListHistory(ctx, productID, skuID, page, pageSize)
}
