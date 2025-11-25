package repository

import (
	"context"
	"github.com/wyfcoding/ecommerce/internal/pricing/domain/entity"
)

// PricingRepository 定价仓储接口
type PricingRepository interface {
	// 规则
	SaveRule(ctx context.Context, rule *entity.PricingRule) error
	GetRule(ctx context.Context, id uint64) (*entity.PricingRule, error)
	GetActiveRule(ctx context.Context, productID, skuID uint64) (*entity.PricingRule, error)
	ListRules(ctx context.Context, productID uint64, offset, limit int) ([]*entity.PricingRule, int64, error)

	// 历史
	SaveHistory(ctx context.Context, history *entity.PriceHistory) error
	ListHistory(ctx context.Context, productID, skuID uint64, offset, limit int) ([]*entity.PriceHistory, int64, error)
}
