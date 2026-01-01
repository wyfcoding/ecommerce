package domain

import (
	"context"
)

// InventoryForecastRepository 是库存预测模块的仓储接口。
type InventoryForecastRepository interface {
	// 销售预测
	SaveForecast(ctx context.Context, forecast *SalesForecast) error
	GetForecastBySKU(ctx context.Context, skuID uint64) (*SalesForecast, error)
	GetSalesHistory(ctx context.Context, skuID uint64, days int) ([]int32, error)

	// 库存预警
	SaveWarning(ctx context.Context, warning *InventoryWarning) error
	ListWarnings(ctx context.Context, offset, limit int) ([]*InventoryWarning, int64, error)

	// 滞销品
	SaveSlowMovingItem(ctx context.Context, item *SlowMovingItem) error
	ListSlowMovingItems(ctx context.Context, offset, limit int) ([]*SlowMovingItem, int64, error)

	// 补货建议
	SaveReplenishmentSuggestion(ctx context.Context, suggestion *ReplenishmentSuggestion) error
	ListReplenishmentSuggestions(ctx context.Context, priority string, offset, limit int) ([]*ReplenishmentSuggestion, int64, error)

	// 缺货风险
	SaveStockoutRisk(ctx context.Context, risk *StockoutRisk) error
	ListStockoutRisks(ctx context.Context, level StockoutRiskLevel, offset, limit int) ([]*StockoutRisk, int64, error)
}
