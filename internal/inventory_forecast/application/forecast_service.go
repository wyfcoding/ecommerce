package application

import (
	"context"

	"github.com/wyfcoding/ecommerce/internal/inventory_forecast/domain"
)

// InventoryForecastService 作为库存预测操作的门面。
type InventoryForecastService struct {
	manager *InventoryForecastManager
	query   *InventoryForecastQuery
}

// NewInventoryForecastService 创建库存预测服务门面实例。
func NewInventoryForecastService(manager *InventoryForecastManager, query *InventoryForecastQuery) *InventoryForecastService {
	return &InventoryForecastService{
		manager: manager,
		query:   query,
	}
}

// --- 写操作（委托给 Manager）---

// GenerateForecast 触发并生成指定SKU的销量预测分析。
func (s *InventoryForecastService) GenerateForecast(ctx context.Context, skuID uint64) (*domain.SalesForecast, error) {
	return s.manager.GenerateForecast(ctx, skuID)
}

// --- 读操作（委托给 Query）---

// GetForecast 获取指定SKU的最新销量预测详情。
func (s *InventoryForecastService) GetForecast(ctx context.Context, skuID uint64) (*domain.SalesForecast, error) {
	return s.query.GetForecast(ctx, skuID)
}

// ListWarnings 获取库存预警记录列表。
func (s *InventoryForecastService) ListWarnings(ctx context.Context, page, pageSize int) ([]*domain.InventoryWarning, int64, error) {
	return s.query.ListWarnings(ctx, page, pageSize)
}

// ListSlowMovingItems 列出当前库存中滞销的商品列表。
func (s *InventoryForecastService) ListSlowMovingItems(ctx context.Context, page, pageSize int) ([]*domain.SlowMovingItem, int64, error) {
	return s.query.ListSlowMovingItems(ctx, page, pageSize)
}

// ListReplenishmentSuggestions 获取针对不同SKU的补货建议列表。
func (s *InventoryForecastService) ListReplenishmentSuggestions(ctx context.Context, priority string, page, pageSize int) ([]*domain.ReplenishmentSuggestion, int64, error) {
	return s.query.ListReplenishmentSuggestions(ctx, priority, page, pageSize)
}

// ListStockoutRisks 获取具有断货风险的SKU列表及其风险等级。
func (s *InventoryForecastService) ListStockoutRisks(ctx context.Context, level domain.StockoutRiskLevel, page, pageSize int) ([]*domain.StockoutRisk, int64, error) {
	return s.query.ListStockoutRisks(ctx, level, page, pageSize)
}
