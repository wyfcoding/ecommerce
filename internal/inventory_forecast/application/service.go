package application

import (
	"context"
	"time"

	"github.com/wyfcoding/ecommerce/internal/inventory_forecast/domain/entity"     // 导入库存预测领域的实体定义。
	"github.com/wyfcoding/ecommerce/internal/inventory_forecast/domain/repository" // 导入库存预测领域的仓储接口。

	"log/slog" // 导入结构化日志库。
)

// InventoryForecastService 结构体定义了库存预测相关的应用服务。
// 它协调领域层和基础设施层，处理销售预测的生成、库存预警、滞销品识别和补货建议等业务逻辑。
type InventoryForecastService struct {
	repo   repository.InventoryForecastRepository // 依赖InventoryForecastRepository接口，用于数据持久化操作。
	logger *slog.Logger                           // 日志记录器，用于记录服务运行时的信息和错误。
}

// NewInventoryForecastService 创建并返回一个新的 InventoryForecastService 实例。
func NewInventoryForecastService(repo repository.InventoryForecastRepository, logger *slog.Logger) *InventoryForecastService {
	return &InventoryForecastService{
		repo:   repo,
		logger: logger,
	}
}

// GenerateForecast 生成销售预测。
// ctx: 上下文。
// skuID: 待预测的SKU ID。
// 返回生成的SalesForecast实体和可能发生的错误。
func (s *InventoryForecastService) GenerateForecast(ctx context.Context, skuID uint64) (*entity.SalesForecast, error) {
	// TODO: 在实际系统中，此处应调用AI模型或使用历史数据进行销售预测。
	// 当前实现创建了一个模拟的销售预测数据。
	forecast := &entity.SalesForecast{
		SKUID:             skuID,
		AverageDailySales: 100,  // 模拟平均日销量。
		TrendRate:         0.05, // 模拟趋势增长率。
		Predictions: []*entity.DailyForecast{ // 模拟每日预测数据。
			{Date: time.Now().AddDate(0, 0, 1), Quantity: 105, Confidence: 0.9},
			{Date: time.Now().AddDate(0, 0, 2), Quantity: 110, Confidence: 0.85},
			{Date: time.Now().AddDate(0, 0, 3), Quantity: 115, Confidence: 0.8},
		},
	}
	// 通过仓储接口保存销售预测。
	if err := s.repo.SaveForecast(ctx, forecast); err != nil {
		s.logger.Error("failed to save forecast", "error", err)
		return nil, err
	}
	return forecast, nil
}

// GetForecast 获取指定SKU的销售预测。
// ctx: 上下文。
// skuID: 待查询的SKU ID。
// 返回SalesForecast实体和可能发生的错误。
func (s *InventoryForecastService) GetForecast(ctx context.Context, skuID uint64) (*entity.SalesForecast, error) {
	return s.repo.GetForecastBySKU(ctx, skuID)
}

// ListWarnings 获取库存预警列表。
// ctx: 上下文。
// page, pageSize: 分页参数。
// 返回库存预警列表、总数和可能发生的错误。
func (s *InventoryForecastService) ListWarnings(ctx context.Context, page, pageSize int) ([]*entity.InventoryWarning, int64, error) {
	offset := (page - 1) * pageSize
	return s.repo.ListWarnings(ctx, offset, pageSize)
}

// ListSlowMovingItems 获取滞销品列表。
// ctx: 上下文。
// page, pageSize: 分页参数。
// 返回滞销品列表、总数和可能发生的错误。
func (s *InventoryForecastService) ListSlowMovingItems(ctx context.Context, page, pageSize int) ([]*entity.SlowMovingItem, int64, error) {
	offset := (page - 1) * pageSize
	return s.repo.ListSlowMovingItems(ctx, offset, pageSize)
}

// ListReplenishmentSuggestions 获取补货建议列表。
// ctx: 上下文。
// priority: 筛选补货建议的优先级。
// page, pageSize: 分页参数。
// 返回补货建议列表、总数和可能发生的错误。
func (s *InventoryForecastService) ListReplenishmentSuggestions(ctx context.Context, priority string, page, pageSize int) ([]*entity.ReplenishmentSuggestion, int64, error) {
	offset := (page - 1) * pageSize
	return s.repo.ListReplenishmentSuggestions(ctx, priority, offset, pageSize)
}

// ListStockoutRisks 获取缺货风险列表。
// ctx: 上下文。
// level: 筛选缺货风险的级别。
// page, pageSize: 分页参数。
// 返回缺货风险列表、总数和可能发生的错误。
func (s *InventoryForecastService) ListStockoutRisks(ctx context.Context, level entity.StockoutRiskLevel, page, pageSize int) ([]*entity.StockoutRisk, int64, error) {
	offset := (page - 1) * pageSize
	return s.repo.ListStockoutRisks(ctx, level, offset, pageSize)
}
