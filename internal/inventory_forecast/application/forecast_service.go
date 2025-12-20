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

// NewInventoryForecastService creates a new InventoryForecastService facade.
func NewInventoryForecastService(manager *InventoryForecastManager, query *InventoryForecastQuery) *InventoryForecastService {
	return &InventoryForecastService{
		manager: manager,
		query:   query,
	}
}

// --- 写操作（委托给 Manager）---

func (s *InventoryForecastService) GenerateForecast(ctx context.Context, skuID uint64) (*domain.SalesForecast, error) {
	return s.manager.GenerateForecast(ctx, skuID)
}

// --- 读操作（委托给 Query）---

func (s *InventoryForecastService) GetForecast(ctx context.Context, skuID uint64) (*domain.SalesForecast, error) {
	return s.query.GetForecast(ctx, skuID)
}

func (s *InventoryForecastService) ListWarnings(ctx context.Context, page, pageSize int) ([]*domain.InventoryWarning, int64, error) {
	return s.query.ListWarnings(ctx, page, pageSize)
}

func (s *InventoryForecastService) ListSlowMovingItems(ctx context.Context, page, pageSize int) ([]*domain.SlowMovingItem, int64, error) {
	return s.query.ListSlowMovingItems(ctx, page, pageSize)
}

func (s *InventoryForecastService) ListReplenishmentSuggestions(ctx context.Context, priority string, page, pageSize int) ([]*domain.ReplenishmentSuggestion, int64, error) {
	return s.query.ListReplenishmentSuggestions(ctx, priority, page, pageSize)
}

func (s *InventoryForecastService) ListStockoutRisks(ctx context.Context, level domain.StockoutRiskLevel, page, pageSize int) ([]*domain.StockoutRisk, int64, error) {
	return s.query.ListStockoutRisks(ctx, level, page, pageSize)
}
