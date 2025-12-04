package repository

import (
	"context"
	"github.com/wyfcoding/ecommerce/internal/dynamic_pricing/domain/entity" // 导入动态定价领域的实体定义。
)

// PricingRepository 是动态定价模块的仓储接口。
// 它定义了对动态价格、竞品价格、价格历史和定价策略实体进行数据持久化操作的契约。
// 仓储接口属于领域层，旨在将领域逻辑与数据存储的实现细节解耦。
type PricingRepository interface {
	// --- DynamicPrice methods ---

	// SaveDynamicPrice 将动态价格实体保存到数据存储中。
	// ctx: 上下文。
	// price: 待保存的动态价格实体。
	SaveDynamicPrice(ctx context.Context, price *entity.DynamicPrice) error
	// GetLatestDynamicPrice 获取指定SKU的最新动态价格实体。
	GetLatestDynamicPrice(ctx context.Context, skuID uint64) (*entity.DynamicPrice, error)
	// ListDynamicPrices 列出指定SKU的所有历史动态价格记录，支持分页。
	ListDynamicPrices(ctx context.Context, skuID uint64, offset, limit int) ([]*entity.DynamicPrice, int64, error)

	// --- CompetitorPrice methods ---

	// SaveCompetitorPriceInfo 将竞品价格汇总信息实体保存到数据存储中。
	SaveCompetitorPriceInfo(ctx context.Context, info *entity.CompetitorPriceInfo) error
	// GetCompetitorPriceInfo 获取指定SKU的竞品价格汇总信息实体。
	GetCompetitorPriceInfo(ctx context.Context, skuID uint64) (*entity.CompetitorPriceInfo, error)

	// --- PriceHistory methods ---

	// SavePriceHistory 将价格历史数据实体保存到数据存储中。
	SavePriceHistory(ctx context.Context, history *entity.PriceHistoryData) error
	// GetPriceHistory 获取指定SKU在过去指定天数内的价格历史数据。
	GetPriceHistory(ctx context.Context, skuID uint64, days int) ([]*entity.PriceHistoryData, error)

	// --- PricingStrategy methods ---

	// SavePricingStrategy 将定价策略实体保存到数据存储中。
	SavePricingStrategy(ctx context.Context, strategy *entity.PricingStrategy) error
	// GetPricingStrategy 获取指定SKU的定价策略实体。
	GetPricingStrategy(ctx context.Context, skuID uint64) (*entity.PricingStrategy, error)
	// ListPricingStrategies 列出所有定价策略实体，支持分页。
	ListPricingStrategies(ctx context.Context, offset, limit int) ([]*entity.PricingStrategy, int64, error)
}
