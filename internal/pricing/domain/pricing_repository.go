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

	// PricingRule (基础定价规则)
	SaveRule(ctx context.Context, rule *PricingRule) error
	SaveRuleInTx(ctx context.Context, tx any, rule *PricingRule) error
	GetRule(ctx context.Context, id uint64) (*PricingRule, error)
	GetActiveRule(ctx context.Context, productID, skuID uint64) (*PricingRule, error)
	ListRules(ctx context.Context, productID uint64, offset, limit int) ([]*PricingRule, int64, error)

	// PriceHistory (价格变动历史)
	SaveHistory(ctx context.Context, history *PriceHistory) error
	SaveHistoryInTx(ctx context.Context, tx any, history *PriceHistory) error
	GetHistory(ctx context.Context, id uint64) (*PriceHistory, error)
	ListHistory(ctx context.Context, productID, skuID uint64, offset, limit int) ([]*PriceHistory, int64, error)

	// DynamicPricing (进阶动态定价)
	SaveDynamicPrice(ctx context.Context, price *DynamicPrice) error
	SaveDynamicPriceInTx(ctx context.Context, tx any, price *DynamicPrice) error
	GetLatestDynamicPrice(ctx context.Context, skuID uint64) (*DynamicPrice, error)

	// Competitor & Elasticity (竞品与弹性分析)
	SaveCompetitorPrice(ctx context.Context, price *CompetitorPrice) error
	SaveCompetitorPriceInTx(ctx context.Context, tx any, price *CompetitorPrice) error
	SaveCompetitorPriceInfo(ctx context.Context, info *CompetitorPriceInfo) error
	SaveCompetitorPriceInfoInTx(ctx context.Context, tx any, info *CompetitorPriceInfo) error
	GetCompetitorPriceInfo(ctx context.Context, skuID uint64) (*CompetitorPriceInfo, error)
	GetPriceElasticity(ctx context.Context, skuID uint64) (*PriceElasticity, error)
	GetDynamicPriceHistory(ctx context.Context, skuID uint64, limit int) ([]DynamicPrice, error)
}
