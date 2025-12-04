package repository

import (
	"context"
	"github.com/wyfcoding/ecommerce/internal/pricing/domain/entity" // 导入定价领域的实体定义。
)

// PricingRepository 是定价模块的仓储接口。
// 它定义了对定价规则和价格历史实体进行数据持久化操作的契约。
// 仓储接口属于领域层，旨在将领域逻辑与数据存储的实现细节解耦。
type PricingRepository interface {
	// --- 定价规则 (PricingRule methods) ---

	// SaveRule 将定价规则实体保存到数据存储中。
	// 如果规则已存在，则更新；如果不存在，则创建。
	// ctx: 上下文。
	// rule: 待保存的定价规则实体。
	SaveRule(ctx context.Context, rule *entity.PricingRule) error
	// GetRule 根据ID获取定价规则实体。
	GetRule(ctx context.Context, id uint64) (*entity.PricingRule, error)
	// GetActiveRule 根据商品ID和SKUID获取当前活跃的定价规则实体。
	GetActiveRule(ctx context.Context, productID, skuID uint64) (*entity.PricingRule, error)
	// ListRules 列出指定商品ID的所有定价规则实体，支持分页。
	ListRules(ctx context.Context, productID uint64, offset, limit int) ([]*entity.PricingRule, int64, error)

	// --- 价格历史 (PriceHistory methods) ---

	// SaveHistory 将价格历史实体保存到数据存储中。
	SaveHistory(ctx context.Context, history *entity.PriceHistory) error
	// ListHistory 列出指定商品ID和SKUID的所有价格历史实体，支持分页。
	ListHistory(ctx context.Context, productID, skuID uint64, offset, limit int) ([]*entity.PriceHistory, int64, error)
}
