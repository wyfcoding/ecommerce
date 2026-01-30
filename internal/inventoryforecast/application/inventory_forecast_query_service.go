package application

import (
	"context"

	"github.com/wyfcoding/ecommerce/internal/inventoryforecast/domain"
)

// InventoryForecastQueryService 处理库存预测的读操作。
type InventoryForecastQueryService struct {
	repo domain.InventoryForecastRepository
}

// NewInventoryForecastQueryService creates a new InventoryForecastQueryService instance.
func NewInventoryForecastQueryService(repo domain.InventoryForecastRepository) *InventoryForecastQueryService {
	return &InventoryForecastQueryService{
		repo: repo,
	}
}

// GetForecast 获取指定SKU的销售预测。
func (q *InventoryForecastQueryService) GetForecast(ctx context.Context, skuID uint64) (*domain.SalesForecast, error) {
	return q.repo.GetForecastBySKU(ctx, skuID)
}

// ListWarnings 获取库存预警列表。
func (q *InventoryForecastQueryService) ListWarnings(ctx context.Context, page, pageSize int) ([]*domain.InventoryWarning, int64, error) {
	offset := (page - 1) * pageSize
	return q.repo.ListWarnings(ctx, offset, pageSize)
}

// ListSlowMovingItems 获取滞销品列表。
func (q *InventoryForecastQueryService) ListSlowMovingItems(ctx context.Context, page, pageSize int) ([]*domain.SlowMovingItem, int64, error) {
	offset := (page - 1) * pageSize
	return q.repo.ListSlowMovingItems(ctx, offset, pageSize)
}

// ListReplenishmentSuggestions 获取补货建议列表。
func (q *InventoryForecastQueryService) ListReplenishmentSuggestions(ctx context.Context, priority string, page, pageSize int) ([]*domain.ReplenishmentSuggestion, int64, error) {
	offset := (page - 1) * pageSize
	return q.repo.ListReplenishmentSuggestions(ctx, priority, offset, pageSize)
}

// ListStockoutRisks 获取缺货风险列表。
func (q *InventoryForecastQueryService) ListStockoutRisks(ctx context.Context, level domain.StockoutRiskLevel, page, pageSize int) ([]*domain.StockoutRisk, int64, error) {
	offset := (page - 1) * pageSize
	return q.repo.ListStockoutRisks(ctx, level, offset, pageSize)
}
