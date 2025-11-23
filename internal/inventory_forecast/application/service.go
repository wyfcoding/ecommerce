package application

import (
	"context"
	"ecommerce/internal/inventory_forecast/domain/entity"
	"ecommerce/internal/inventory_forecast/domain/repository"
	"time"

	"log/slog"
)

type InventoryForecastService struct {
	repo   repository.InventoryForecastRepository
	logger *slog.Logger
}

func NewInventoryForecastService(repo repository.InventoryForecastRepository, logger *slog.Logger) *InventoryForecastService {
	return &InventoryForecastService{
		repo:   repo,
		logger: logger,
	}
}

// GenerateForecast 生成销售预测 (Mock logic)
func (s *InventoryForecastService) GenerateForecast(ctx context.Context, skuID uint64) (*entity.SalesForecast, error) {
	// In a real system, this would call an AI model or use historical data.
	// Here we just create a mock forecast.
	forecast := &entity.SalesForecast{
		SKUID:             skuID,
		AverageDailySales: 100,
		TrendRate:         0.05,
		Predictions: []*entity.DailyForecast{
			{Date: time.Now().AddDate(0, 0, 1), Quantity: 105, Confidence: 0.9},
			{Date: time.Now().AddDate(0, 0, 2), Quantity: 110, Confidence: 0.85},
			{Date: time.Now().AddDate(0, 0, 3), Quantity: 115, Confidence: 0.8},
		},
	}
	if err := s.repo.SaveForecast(ctx, forecast); err != nil {
		s.logger.Error("failed to save forecast", "error", err)
		return nil, err
	}
	return forecast, nil
}

// GetForecast 获取预测
func (s *InventoryForecastService) GetForecast(ctx context.Context, skuID uint64) (*entity.SalesForecast, error) {
	return s.repo.GetForecastBySKU(ctx, skuID)
}

// ListWarnings 获取预警列表
func (s *InventoryForecastService) ListWarnings(ctx context.Context, page, pageSize int) ([]*entity.InventoryWarning, int64, error) {
	offset := (page - 1) * pageSize
	return s.repo.ListWarnings(ctx, offset, pageSize)
}

// ListSlowMovingItems 获取滞销品列表
func (s *InventoryForecastService) ListSlowMovingItems(ctx context.Context, page, pageSize int) ([]*entity.SlowMovingItem, int64, error) {
	offset := (page - 1) * pageSize
	return s.repo.ListSlowMovingItems(ctx, offset, pageSize)
}

// ListReplenishmentSuggestions 获取补货建议
func (s *InventoryForecastService) ListReplenishmentSuggestions(ctx context.Context, priority string, page, pageSize int) ([]*entity.ReplenishmentSuggestion, int64, error) {
	offset := (page - 1) * pageSize
	return s.repo.ListReplenishmentSuggestions(ctx, priority, offset, pageSize)
}

// ListStockoutRisks 获取缺货风险
func (s *InventoryForecastService) ListStockoutRisks(ctx context.Context, level entity.StockoutRiskLevel, page, pageSize int) ([]*entity.StockoutRisk, int64, error) {
	offset := (page - 1) * pageSize
	return s.repo.ListStockoutRisks(ctx, level, offset, pageSize)
}
