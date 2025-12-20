package domain

import (
	"context"
)

// PricingRepository 是动态定价模块的仓储接口。
type PricingRepository interface {
	// --- DynamicPrice methods ---
	SaveDynamicPrice(ctx context.Context, price *DynamicPrice) error
	GetLatestDynamicPrice(ctx context.Context, skuID uint64) (*DynamicPrice, error)
	ListDynamicPrices(ctx context.Context, skuID uint64, offset, limit int) ([]*DynamicPrice, int64, error)

	// --- CompetitorPrice methods ---
	SaveCompetitorPriceInfo(ctx context.Context, info *CompetitorPriceInfo) error
	GetCompetitorPriceInfo(ctx context.Context, skuID uint64) (*CompetitorPriceInfo, error)

	// --- PriceHistory methods ---
	SavePriceHistory(ctx context.Context, history *PriceHistoryData) error
	GetPriceHistory(ctx context.Context, skuID uint64, days int) ([]*PriceHistoryData, error)

	// --- PricingStrategy methods ---
	SavePricingStrategy(ctx context.Context, strategy *PricingStrategy) error
	GetPricingStrategy(ctx context.Context, skuID uint64) (*PricingStrategy, error)
	ListPricingStrategies(ctx context.Context, offset, limit int) ([]*PricingStrategy, int64, error)
}
