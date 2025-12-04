package repository

import (
	"context"
	"github.com/wyfcoding/ecommerce/internal/inventory_forecast/domain/entity" // 导入库存预测领域的实体定义。
)

// InventoryForecastRepository 是库存预测模块的仓储接口。
// 它定义了对销售预测、库存预警、滞销品、补货建议和缺货风险实体进行数据持久化操作的契约。
// 仓储接口属于领域层，旨在将领域逻辑与数据存储的实现细节解耦。
type InventoryForecastRepository interface {
	// --- 销售预测 (SalesForecast methods) ---

	// SaveForecast 将销售预测实体保存到数据存储中。
	// ctx: 上下文。
	// forecast: 待保存的销售预测实体。
	SaveForecast(ctx context.Context, forecast *entity.SalesForecast) error
	// GetForecastBySKU 根据SKU ID获取销售预测实体。
	GetForecastBySKU(ctx context.Context, skuID uint64) (*entity.SalesForecast, error)

	// --- 库存预警 (InventoryWarning methods) ---

	// SaveWarning 将库存预警实体保存到数据存储中。
	SaveWarning(ctx context.Context, warning *entity.InventoryWarning) error
	// ListWarnings 列出所有库存预警实体，支持分页。
	ListWarnings(ctx context.Context, offset, limit int) ([]*entity.InventoryWarning, int64, error)

	// --- 滞销品 (SlowMovingItem methods) ---

	// SaveSlowMovingItem 将滞销品实体保存到数据存储中。
	SaveSlowMovingItem(ctx context.Context, item *entity.SlowMovingItem) error
	// ListSlowMovingItems 列出所有滞销品实体，支持分页。
	ListSlowMovingItems(ctx context.Context, offset, limit int) ([]*entity.SlowMovingItem, int64, error)

	// --- 补货建议 (ReplenishmentSuggestion methods) ---

	// SaveReplenishmentSuggestion 将补货建议实体保存到数据存储中。
	SaveReplenishmentSuggestion(ctx context.Context, suggestion *entity.ReplenishmentSuggestion) error
	// ListReplenishmentSuggestions 列出所有补货建议实体，支持通过优先级过滤和分页。
	ListReplenishmentSuggestions(ctx context.Context, priority string, offset, limit int) ([]*entity.ReplenishmentSuggestion, int64, error)

	// --- 缺货风险 (StockoutRisk methods) ---

	// SaveStockoutRisk 将缺货风险实体保存到数据存储中。
	SaveStockoutRisk(ctx context.Context, risk *entity.StockoutRisk) error
	// ListStockoutRisks 列出所有缺货风险实体，支持通过风险等级过滤和分页。
	ListStockoutRisks(ctx context.Context, level entity.StockoutRiskLevel, offset, limit int) ([]*entity.StockoutRisk, int64, error)
}
