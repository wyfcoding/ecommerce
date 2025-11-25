package repository

import (
	"context"
	"github.com/wyfcoding/ecommerce/internal/dynamic_pricing/domain/entity"
)

// PricingRepository 定价仓储接口
type PricingRepository interface {
	// DynamicPrice methods
	SaveDynamicPrice(ctx context.Context, price *entity.DynamicPrice) error
	GetLatestDynamicPrice(ctx context.Context, skuID uint64) (*entity.DynamicPrice, error)
	ListDynamicPrices(ctx context.Context, skuID uint64, offset, limit int) ([]*entity.DynamicPrice, int64, error)

	// CompetitorPrice methods
	SaveCompetitorPriceInfo(ctx context.Context, info *entity.CompetitorPriceInfo) error
	GetCompetitorPriceInfo(ctx context.Context, skuID uint64) (*entity.CompetitorPriceInfo, error)

	// PriceHistory methods
	SavePriceHistory(ctx context.Context, history *entity.PriceHistoryData) error
	GetPriceHistory(ctx context.Context, skuID uint64, days int) ([]*entity.PriceHistoryData, error)

	// PricingStrategy methods
	SavePricingStrategy(ctx context.Context, strategy *entity.PricingStrategy) error
	GetPricingStrategy(ctx context.Context, skuID uint64) (*entity.PricingStrategy, error)
	ListPricingStrategies(ctx context.Context, offset, limit int) ([]*entity.PricingStrategy, int64, error)
}
