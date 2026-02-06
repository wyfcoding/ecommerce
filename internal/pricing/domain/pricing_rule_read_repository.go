// 生成摘要：定义定价规则读模型仓储接口（Redis）。
package domain

import "context"

// PricingRuleReadRepository 定义定价规则缓存接口。
type PricingRuleReadRepository interface {
	Save(ctx context.Context, rule *PricingRule) error
	GetByID(ctx context.Context, id uint64) (*PricingRule, error)
	GetActive(ctx context.Context, productID, skuID uint64) (*PricingRule, error)
	Delete(ctx context.Context, id uint64) error
}
