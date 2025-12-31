package application

import (
	"context"
	"log/slog"
	"time"

	"github.com/wyfcoding/ecommerce/internal/dynamicpricing/domain"
	"github.com/wyfcoding/pkg/algorithm"
)

// DynamicPricingManager 处理动态定价的写操作。
type DynamicPricingManager struct {
	repo   domain.PricingRepository
	logger *slog.Logger
}

// NewDynamicPricingManager 创建并返回一个新的 DynamicPricingManager 实例。
func NewDynamicPricingManager(repo domain.PricingRepository, logger *slog.Logger) *DynamicPricingManager {
	return &DynamicPricingManager{
		repo:   repo,
		logger: logger,
	}
}

// CalculatePrice 计算给定SKU的动态价格。
func (m *DynamicPricingManager) CalculatePrice(ctx context.Context, req *domain.PricingRequest) (*domain.DynamicPrice, error) {
	strategy, err := m.repo.GetPricingStrategy(ctx, req.SKUID)
	if err != nil {
		strategy = &domain.PricingStrategy{
			StrategyType: "fixed",
			MinPrice:     req.BasePrice,
			MaxPrice:     req.BasePrice,
		}
	}

	// 1. 准备定价因素
	now := time.Now()
	demandLevel := 0.5
	if req.AverageDailyDemand > 0 {
		demandLevel = 0.5 * float64(req.DailyDemand) / float64(req.AverageDailyDemand)
	}

	userLevel := 1
	switch req.UserLevel {
	case "VIP":
		userLevel = 9
	case "Diamond":
		userLevel = 10
	case "Gold":
		userLevel = 7
	case "Silver":
		userLevel = 5
	default:
		userLevel = 1
	}

	factors := algorithm.PricingFactors{
		Stock:           req.CurrentStock,
		TotalStock:      req.TotalStock,
		DemandLevel:     demandLevel,
		CompetitorPrice: req.CompetitorPrice,
		TimeOfDay:       now.Hour(),
		DayOfWeek:       int(now.Weekday()),
		IsHoliday:       false, // TODO: 对接日历服务
		UserLevel:       userLevel,
		SeasonFactor:    0.5, // 默认平均季节因素
	}

	// 2. 初始化定价引擎
	// 默认弹性系数 1.0，如果有历史数据可以计算更准确的值
	elasticity := 1.0
	minPrice := strategy.MinPrice
	if minPrice == 0 {
		minPrice = int64(float64(req.BasePrice) * 0.5) // 默认最低半价
	}
	maxPrice := strategy.MaxPrice
	if maxPrice == 0 {
		maxPrice = int64(float64(req.BasePrice) * 2.0) // 默认最高双倍
	}

	engine := algorithm.NewPricingEngine(req.BasePrice, minPrice, maxPrice, elasticity)

	// 3. 计算价格
	result := engine.CalculatePrice(factors)

	// 4. 构建结果
	// Calculate adjustment ratio
	adjustment := 1.0
	if req.BasePrice > 0 {
		adjustment = float64(result.FinalPrice) / float64(req.BasePrice)
	}

	price := &domain.DynamicPrice{
		SKUID:            req.SKUID,
		BasePrice:        req.BasePrice,
		FinalPrice:       result.FinalPrice,
		PriceAdjustment:  adjustment,
		InventoryFactor:  result.InventoryFactor,
		DemandFactor:     result.DemandFactor,
		CompetitorFactor: result.CompetitorFactor,
		TimeFactor:       result.TimeFactor,
		UserFactor:       result.UserFactor,
		EffectiveTime:    now,
		ExpiryTime:       now.Add(24 * time.Hour),
	}

	if err := m.repo.SaveDynamicPrice(ctx, price); err != nil {
		m.logger.ErrorContext(ctx, "failed to save dynamic price", "sku_id", req.SKUID, "error", err)
		return nil, err
	}
	m.logger.InfoContext(ctx, "dynamic price calculated successfully", "sku_id", req.SKUID, "final_price", result.FinalPrice)

	return price, nil
}

// SaveStrategy 保存（创建或更新）一个定价策略。
func (m *DynamicPricingManager) SaveStrategy(ctx context.Context, strategy *domain.PricingStrategy) error {
	return m.repo.SavePricingStrategy(ctx, strategy)
}
