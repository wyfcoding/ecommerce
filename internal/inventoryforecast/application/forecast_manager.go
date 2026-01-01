package application

import (
	"context"
	"log/slog"
	"math"
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
	// 1. 模拟获取历史销量数据
	history := []float64{95, 105, 90, 110, 120, 85, 100, 115}

	// 2. 初始化卡尔曼滤波器和 EWMA (用于计算增长趋势)
	kf := algorithm.NewKalmanFilter(0.01, 0.1, history[0])
	ewmaTrend := algorithm.NewEWMA(0.2)

	var currentLevel float64
	var lastLevel float64
	for i, val := range history {
		lastLevel = currentLevel
		currentLevel = kf.Update(val)

		if i > 0 {
			// 计算环比增长趋势并使用 EWMA 平滑
			growth := (currentLevel - lastLevel) / lastLevel
			ewmaTrend.Update(growth)
		}
	}

	trendRate := ewmaTrend.Value()
	confidence := kf.GetConfidence()

	forecast := &domain.SalesForecast{
		SKUID:             skuID,
		AverageDailySales: int32(currentLevel),
		TrendRate:         trendRate,
		Predictions:       make(domain.DailyForecastArray, 0),
	}

	// 3. 生成未来 3 天的动态预测（考虑趋势和置信度衰减）
	for i := 1; i <= 3; i++ {
		// 预测值 = 当前水平 * (1 + 趋势)^天数
		predictedQty := currentLevel * math.Pow(1+trendRate, float64(i))

		// 修复：DailyForecastArray 是 []*DailyForecast，需要传指针
		forecast.Predictions = append(forecast.Predictions, &domain.DailyForecast{
			Date:       time.Now().AddDate(0, 0, i),
			Quantity:   int32(predictedQty),
			Confidence: confidence * math.Pow(0.9, float64(i-1)), // 置信度随时间推移逐渐降低
		})
	}

	if err := m.repo.SaveForecast(ctx, forecast); err != nil {
		m.logger.ErrorContext(ctx, "failed to save forecast", "sku_id", skuID, "error", err)
		return nil, err
	}

	m.logger.InfoContext(ctx, "advanced forecast generated",
		"sku_id", skuID,
		"smoothed_level", currentLevel,
		"trend", trendRate,
		"confidence", confidence)

	return forecast, nil
}
