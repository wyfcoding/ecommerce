package application

import (
	"context"
	"log/slog"
	"time"

	"github.com/wyfcoding/ecommerce/internal/inventoryforecast/domain"
	"github.com/wyfcoding/pkg/algorithm"
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

// SmoothSalesData 使用卡尔曼滤波平滑历史销量数据
func (m *InventoryForecastManager) SmoothSalesData(history []float64) float64 {
	if len(history) == 0 {
		return 0
	}

	// Q=0.01 (系统稳定度), R=0.1 (对测量值的信任度)
	kf := algorithm.NewKalmanFilter(0.01, 0.1, history[0])

	var smoothVal float64
	for _, val := range history {
		smoothVal = kf.Update(val)
	}
	return smoothVal
}

// GenerateForecast 生成销售预测。
func (m *InventoryForecastManager) GenerateForecast(ctx context.Context, skuID uint64) (*domain.SalesForecast, error) {
	// 模拟从 Repo 获取历史销量数据
	// history, _ := m.repo.GetHistoricalSales(ctx, skuID, 30)
	history := []float64{95, 105, 90, 110, 120, 85, 100, 115} // 示例波动数据

	// 使用卡尔曼滤波得到平滑后的当前销售水平
	currentLevel := m.SmoothSalesData(history)

	forecast := &domain.SalesForecast{
		SKUID:             skuID,
		AverageDailySales: int32(currentLevel),
		TrendRate:         0.05,
		Predictions: domain.DailyForecastArray{
			{Date: time.Now().AddDate(0, 0, 1), Quantity: int32(currentLevel * 1.05), Confidence: 0.9},
			{Date: time.Now().AddDate(0, 0, 2), Quantity: int32(currentLevel * 1.10), Confidence: 0.85},
			{Date: time.Now().AddDate(0, 0, 3), Quantity: int32(currentLevel * 1.15), Confidence: 0.8},
		},
	}
	if err := m.repo.SaveForecast(ctx, forecast); err != nil {
		m.logger.ErrorContext(ctx, "failed to save forecast", "sku_id", skuID, "error", err)
		return nil, err
	}
	m.logger.InfoContext(ctx, "forecast generated successfully", "sku_id", skuID, "smoothed_level", currentLevel)
	return forecast, nil
}
