package repository

import (
	"context"
	"github.com/wyfcoding/ecommerce/internal/inventory_forecast/domain/entity"
)

// InventoryForecastRepository 库存预测仓储接口
type InventoryForecastRepository interface {
	// 销售预测
	SaveForecast(ctx context.Context, forecast *entity.SalesForecast) error
	GetForecastBySKU(ctx context.Context, skuID uint64) (*entity.SalesForecast, error)

	// 库存预警
	SaveWarning(ctx context.Context, warning *entity.InventoryWarning) error
	ListWarnings(ctx context.Context, offset, limit int) ([]*entity.InventoryWarning, int64, error)

	// 滞销品
	SaveSlowMovingItem(ctx context.Context, item *entity.SlowMovingItem) error
	ListSlowMovingItems(ctx context.Context, offset, limit int) ([]*entity.SlowMovingItem, int64, error)

	// 补货建议
	SaveReplenishmentSuggestion(ctx context.Context, suggestion *entity.ReplenishmentSuggestion) error
	ListReplenishmentSuggestions(ctx context.Context, priority string, offset, limit int) ([]*entity.ReplenishmentSuggestion, int64, error)

	// 缺货风险
	SaveStockoutRisk(ctx context.Context, risk *entity.StockoutRisk) error
	ListStockoutRisks(ctx context.Context, level entity.StockoutRiskLevel, offset, limit int) ([]*entity.StockoutRisk, int64, error)
}
