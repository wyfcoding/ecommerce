package application

import (
	"context"
	"log/slog"

	"github.com/wyfcoding/ecommerce/internal/dynamicpricing/domain"
)

// DynamicPricingQueryService 处理动态定价的读操作。
type DynamicPricingQueryService struct {
	repo           domain.PricingRepository
	priceRead      domain.DynamicPriceReadRepository
	strategyRead   domain.PricingStrategyReadRepository
	strategySearch domain.PricingStrategySearchRepository
	logger         *slog.Logger
}

// NewDynamicPricingQueryService 创建并返回一个新的 DynamicPricingQueryService 实例。
func NewDynamicPricingQueryService(
	repo domain.PricingRepository,
	priceRead domain.DynamicPriceReadRepository,
	strategyRead domain.PricingStrategyReadRepository,
	strategySearch domain.PricingStrategySearchRepository,
	logger *slog.Logger,
) *DynamicPricingQueryService {
	return &DynamicPricingQueryService{
		repo:           repo,
		priceRead:      priceRead,
		strategyRead:   strategyRead,
		strategySearch: strategySearch,
		logger:         logger,
	}
}

// GetLatestPrice 获取指定SKU的最新动态价格。
func (q *DynamicPricingQueryService) GetLatestPrice(ctx context.Context, skuID uint64) (*domain.DynamicPrice, error) {
	if q.priceRead != nil {
		if cached, err := q.priceRead.GetLatest(ctx, skuID); err == nil && cached != nil {
			return cached, nil
		}
	}
	price, err := q.repo.GetLatestDynamicPrice(ctx, skuID)
	if err != nil {
		return nil, err
	}
	if price != nil && q.priceRead != nil {
		_ = q.priceRead.SaveLatest(ctx, price)
	}
	return price, nil
}

// ListStrategies 获取定价策略列表。
func (q *DynamicPricingQueryService) ListStrategies(ctx context.Context, page, pageSize int) ([]*domain.PricingStrategy, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	query := &domain.PricingStrategyQuery{
		Page:     page,
		PageSize: pageSize,
	}
	offset := (page - 1) * pageSize
	if q.strategySearch != nil {
		list, total, err := q.strategySearch.Search(ctx, query, offset, pageSize)
		if err == nil {
			return list, total, nil
		}
		if q.logger != nil {
			q.logger.WarnContext(ctx, "pricing strategy search fallback to mysql", "error", err)
		}
	}
	return q.repo.ListPricingStrategies(ctx, query)
}

// GetPricingStrategy 获取指定SKU的定价策略。
func (q *DynamicPricingQueryService) GetPricingStrategy(ctx context.Context, skuID uint64) (*domain.PricingStrategy, error) {
	if q.strategyRead != nil {
		if cached, err := q.strategyRead.GetBySKU(ctx, skuID); err == nil && cached != nil {
			return cached, nil
		}
	}
	strategy, err := q.repo.GetPricingStrategy(ctx, skuID)
	if err != nil {
		return nil, err
	}
	if strategy != nil && q.strategyRead != nil {
		_ = q.strategyRead.Save(ctx, strategy)
	}
	return strategy, nil
}
