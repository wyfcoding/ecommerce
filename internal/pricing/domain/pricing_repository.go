package domain

import (
	"context"
)

// PricingRepository 是定价模块的仓储接口。
type PricingRepository interface {
	// 事务支持
	BeginTx(ctx context.Context) any
	CommitTx(tx any) error
	RollbackTx(tx any) error
	WithTx(ctx context.Context, fn func(tx any) error) error

	// PricingRule
	SaveRule(ctx context.Context, rule *PricingRule) error
	SaveRuleInTx(ctx context.Context, tx any, rule *PricingRule) error
	GetRule(ctx context.Context, id uint64) (*PricingRule, error)
	GetActiveRule(ctx context.Context, productID, skuID uint64) (*PricingRule, error)
	ListRules(ctx context.Context, productID uint64, offset, limit int) ([]*PricingRule, int64, error)

	// PriceHistory
	SaveHistory(ctx context.Context, history *PriceHistory) error
	SaveHistoryInTx(ctx context.Context, tx any, history *PriceHistory) error
	GetHistory(ctx context.Context, id uint64) (*PriceHistory, error)
	ListHistory(ctx context.Context, productID, skuID uint64, offset, limit int) ([]*PriceHistory, int64, error)
}
