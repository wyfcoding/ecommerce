package domain

import "context"

// PricingRepository 定义了动态定价模块的数据持久层接口。
type PricingRepository interface {
	// --- tx helpers ---
	BeginTx(ctx context.Context) any
	CommitTx(tx any) error
	RollbackTx(tx any) error
	WithTx(ctx context.Context, fn func(tx any) error) error

	// 动态价格
	SaveDynamicPrice(ctx context.Context, price *DynamicPrice) error
	SaveDynamicPriceInTx(ctx context.Context, tx any, price *DynamicPrice) error
	GetLatestDynamicPrice(ctx context.Context, skuID uint64) (*DynamicPrice, error)

	// 定价策略
	SavePricingStrategy(ctx context.Context, strategy *PricingStrategy) error
	SavePricingStrategyInTx(ctx context.Context, tx any, strategy *PricingStrategy) error
	GetPricingStrategy(ctx context.Context, skuID uint64) (*PricingStrategy, error)
	ListPricingStrategies(ctx context.Context, query *PricingStrategyQuery) ([]*PricingStrategy, int64, error)

	// 竞品与弹性数据
	SaveCompetitorPrice(ctx context.Context, price *CompetitorPrice) error
	SaveCompetitorPriceInTx(ctx context.Context, tx any, price *CompetitorPrice) error
	SaveCompetitorPriceInfo(ctx context.Context, info *CompetitorPriceInfo) error
	SaveCompetitorPriceInfoInTx(ctx context.Context, tx any, info *CompetitorPriceInfo) error
	GetPriceElasticity(ctx context.Context, skuID uint64) (*PriceElasticity, error)
	GetCompetitorPriceInfo(ctx context.Context, skuID uint64) (*CompetitorPriceInfo, error)
	GetPriceHistory(ctx context.Context, skuID uint64, limit int) ([]PriceHistoryData, error)
}

// PricingStrategyQuery 定价策略查询条件。
type PricingStrategyQuery struct {
	SKUID    uint64
	Enabled  *bool
	Page     int
	PageSize int
}
