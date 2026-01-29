package application

import (
	"context"

	"github.com/wyfcoding/ecommerce/internal/dynamicpricing/domain"
)

// DynamicPricingService 作为动态定价操作的门面。
type DynamicPricingService struct {
	Command *DynamicPricingCommandService
	Query   *DynamicPricingQueryService
}

// NewDynamicPricingService 创建动态定价服务门面实例。
func NewDynamicPricingService(command *DynamicPricingCommandService, query *DynamicPricingQueryService) *DynamicPricingService {
	return &DynamicPricingService{
		Command: command,
		Query:   query,
	}
}

// --- 写操作（委托给 Command）---

// CalculatePrice 核心算法：基于实时策略计算动态价格。
func (s *DynamicPricingService) CalculatePrice(ctx context.Context, req *domain.PricingRequest) (*domain.DynamicPrice, error) {
	return s.Command.CalculatePrice(ctx, req)
}

// SaveStrategy 保存或更新动态定价策略配置。
func (s *DynamicPricingService) SaveStrategy(ctx context.Context, strategy *domain.PricingStrategy) error {
	return s.Command.SaveStrategy(ctx, strategy)
}

// --- 读操作（委托给 Query）---

// GetLatestPrice 获取指定SKU的最新动态价格快照。
func (s *DynamicPricingService) GetLatestPrice(ctx context.Context, skuID uint64) (*domain.DynamicPrice, error) {
	return s.Query.GetLatestPrice(ctx, skuID)
}

// ListStrategies 获取所有定价策略列表（分页）。
func (s *DynamicPricingService) ListStrategies(ctx context.Context, page, pageSize int) ([]*domain.PricingStrategy, int64, error) {
	return s.Query.ListStrategies(ctx, page, pageSize)
}
