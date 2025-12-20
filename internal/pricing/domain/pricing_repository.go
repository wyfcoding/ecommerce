package domain

import (
	"context"
)

// PricingRepository 是定价模块的仓储接口。
type PricingRepository interface {
	// PricingRule
	SaveRule(ctx context.Context, rule *PricingRule) error
	GetRule(ctx context.Context, id uint64) (*PricingRule, error)
	GetActiveRule(ctx context.Context, productID, skuID uint64) (*PricingRule, error)
	ListRules(ctx context.Context, productID uint64, offset, limit int) ([]*PricingRule, int64, error)

	// PriceHistory
	SaveHistory(ctx context.Context, history *PriceHistory) error
	ListHistory(ctx context.Context, productID, skuID uint64, offset, limit int) ([]*PriceHistory, int64, error)
}
