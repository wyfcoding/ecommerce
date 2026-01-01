package domain

import (
	"context"
)

// PricingRepository 定义了动态定价模块的数据持久层接口。
type PricingRepository interface {
	// 动态价格
	SaveDynamicPrice(ctx context.Context, price *DynamicPrice) error
	GetLatestDynamicPrice(ctx context.Context, skuID uint64) (*DynamicPrice, error)

	// 定价策略
	SavePricingStrategy(ctx context.Context, strategy *PricingStrategy) error
	GetPricingStrategy(ctx context.Context, skuID uint64) (*PricingStrategy, error)
	ListPricingStrategies(ctx context.Context, offset, limit int) ([]*PricingStrategy, int64, error)

	// 高级数据获取
	GetPriceElasticity(ctx context.Context, skuID uint64) (*PriceElasticity, error)
	GetCompetitorPriceInfo(ctx context.Context, skuID uint64) (*CompetitorPriceInfo, error)
	GetPriceHistory(ctx context.Context, skuID uint64, limit int) ([]PriceHistoryData, error)
}
