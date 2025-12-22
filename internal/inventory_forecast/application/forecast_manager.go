package application

import (
	"context"
	"log/slog"
	"time"

	"github.com/wyfcoding/ecommerce/internal/inventory_forecast/domain"
)

// InventoryForecastManager 处理库存预测的写操作。
type InventoryForecastManager struct {
	repo   domain.InventoryForecastRepository
	logger *slog.Logger
}

// NewInventoryForecastManager creates a new InventoryForecastManager instance.
func NewInventoryForecastManager(repo domain.InventoryForecastRepository, logger *slog.Logger) *InventoryForecastManager {
	return &InventoryForecastManager{
		repo:   repo,
		logger: logger,
	}
}

// GenerateForecast 生成销售预测。
func (m *InventoryForecastManager) GenerateForecast(ctx context.Context, skuID uint64) (*domain.SalesForecast, error) {
	forecast := &domain.SalesForecast{
		SKUID:             skuID,
		AverageDailySales: 100,
		TrendRate:         0.05,
		Predictions: []*domain.DailyForecast{
			{Date: time.Now().AddDate(0, 0, 1), Quantity: 105, Confidence: 0.9},
			{Date: time.Now().AddDate(0, 0, 2), Quantity: 110, Confidence: 0.85},
			{Date: time.Now().AddDate(0, 0, 3), Quantity: 115, Confidence: 0.8},
		},
	}
	if err := m.repo.SaveForecast(ctx, forecast); err != nil {
		m.logger.ErrorContext(ctx, "failed to save forecast", "sku_id", skuID, "error", err)
		return nil, err
	}
	m.logger.InfoContext(ctx, "forecast generated successfully", "sku_id", skuID)
	return forecast, nil
}
