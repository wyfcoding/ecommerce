package domain

import "context"

// PricingStrategySearchRepository 定义定价策略搜索仓储接口（Elasticsearch）。
type PricingStrategySearchRepository interface {
	Index(ctx context.Context, strategy *PricingStrategy) error
	Delete(ctx context.Context, skuID uint64) error
	Search(ctx context.Context, query *PricingStrategyQuery, offset, limit int) ([]*PricingStrategy, int64, error)
}
