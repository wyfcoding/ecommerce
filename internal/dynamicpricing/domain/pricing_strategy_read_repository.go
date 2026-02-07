package domain

import "context"

// PricingStrategyReadRepository 定义定价策略读模型仓储接口（Redis）。
type PricingStrategyReadRepository interface {
	Save(ctx context.Context, strategy *PricingStrategy) error
	GetBySKU(ctx context.Context, skuID uint64) (*PricingStrategy, error)
	Delete(ctx context.Context, skuID uint64) error
}
