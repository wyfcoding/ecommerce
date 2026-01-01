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
	// 1. 获取基础策略
	strategy, err := m.repo.GetPricingStrategy(ctx, req.SKUID)
	if err != nil || strategy == nil {
		strategy = &domain.PricingStrategy{
			StrategyType: "dynamic", // Default to dynamic if not found
			MinPrice:     int64(float64(req.BasePrice) * 0.5),
			MaxPrice:     int64(float64(req.BasePrice) * 2.0),
		}
	}

	// 2. 获取价格弹性 (Elasticity)
	elasticityVal := 1.0
	if el, err := m.repo.GetPriceElasticity(ctx, req.SKUID); err == nil && el != nil {
		elasticityVal = el.Elasticity
	}

	// 3. 初始化定价引擎
	minPrice := strategy.MinPrice
	if minPrice == 0 {
		minPrice = int64(float64(req.BasePrice) * 0.5)
	}
	maxPrice := strategy.MaxPrice
	if maxPrice == 0 {
		maxPrice = int64(float64(req.BasePrice) * 2.0)
	}
	engine := algorithm.NewPricingEngine(req.BasePrice, minPrice, maxPrice, elasticityVal)

	var finalPrice int64
	var algoFactors domain.DynamicPrice // Reuse struct to store factors

	// 4. 根据策略类型执行计算
	switch strategy.StrategyType {
	case "profit_maximization":
		// 利润最大化策略
		// 获取历史数据用于需求预测
		history, _ := m.repo.GetPriceHistory(ctx, req.SKUID, 30) // 最近30条记录
		
		// 转换历史数据格式
		var demandData []algorithm.DemandData
		for _, h := range history {
			demandData = append(demandData, algorithm.DemandData{
				Price:  h.Price,
				Demand: int64(h.Quantity),
			})
		}

		// 定义需求预测函数
		demandFunc := func(p int64) int64 {
			return engine.PredictDemand(p, demandData)
		}

		// 估算成本 (假设 30% 毛利空间，即成本为 BasePrice * 0.7)
		cost := int64(float64(req.BasePrice) * 0.7)
		finalPrice = engine.OptimalPriceForProfit(cost, demandFunc)

		// 填充默认因子
		algoFactors.DemandFactor = 1.0 
		algoFactors.InventoryFactor = 1.0

	case "competitive":
		// 竞争定价策略
		compInfo, _ := m.repo.GetCompetitorPriceInfo(ctx, req.SKUID)
		var compPrices []int64
		if compInfo != nil {
			for _, c := range compInfo.Competitors {
				compPrices = append(compPrices, c.Price)
			}
		}
		// 使用 "average" 跟随策略 (也可以从 strategy 配置中读取子策略)
		finalPrice = engine.CompetitivePricing(compPrices, "average")
		algoFactors.CompetitorFactor = float64(finalPrice) / float64(req.BasePrice)

	default:
		// 默认 dynamic 策略 (综合因子)
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
		}

		// 尝试获取竞品价格
		compPrice := req.CompetitorPrice
		if compPrice == 0 {
			if compInfo, err := m.repo.GetCompetitorPriceInfo(ctx, req.SKUID); err == nil && compInfo != nil {
				compPrice = compInfo.LowestPrice // 使用最低竞品价作为参考
			}
		}

		factors := algorithm.PricingFactors{
			Stock:           req.CurrentStock,
			TotalStock:      req.TotalStock,
			DemandLevel:     demandLevel,
			CompetitorPrice: compPrice,
			TimeOfDay:       now.Hour(),
			DayOfWeek:       int(now.Weekday()),
			IsHoliday:       false, // TODO: 对接日历服务
			UserLevel:       userLevel,
			SeasonFactor:    0.5,
		}

		result := engine.CalculatePrice(factors)
		finalPrice = result.FinalPrice
		
		// 记录因子
		algoFactors.InventoryFactor = result.InventoryFactor
		algoFactors.DemandFactor = result.DemandFactor
		algoFactors.CompetitorFactor = result.CompetitorFactor
		algoFactors.TimeFactor = result.TimeFactor
		algoFactors.UserFactor = result.UserFactor
	}

	// 5. 构建并保存结果
	adjustment := 1.0
	if req.BasePrice > 0 {
		adjustment = float64(finalPrice) / float64(req.BasePrice)
	}

	price := &domain.DynamicPrice{
		SKUID:            req.SKUID,
		BasePrice:        req.BasePrice,
		FinalPrice:       finalPrice,
		PriceAdjustment:  adjustment,
		InventoryFactor:  algoFactors.InventoryFactor,
		DemandFactor:     algoFactors.DemandFactor,
		CompetitorFactor: algoFactors.CompetitorFactor,
		TimeFactor:       algoFactors.TimeFactor,
		UserFactor:       algoFactors.UserFactor,
		EffectiveTime:    time.Now(),
		ExpiryTime:       time.Now().Add(24 * time.Hour),
	}

	if err := m.repo.SaveDynamicPrice(ctx, price); err != nil {
		m.logger.ErrorContext(ctx, "failed to save dynamic price", "sku_id", req.SKUID, "error", err)
		return nil, err
	}
	m.logger.InfoContext(ctx, "dynamic price calculated successfully", "sku_id", req.SKUID, "strategy", strategy.StrategyType, "final_price", finalPrice)

	return price, nil
}

// SaveStrategy 保存（创建或更新）一个定价策略。
func (m *DynamicPricingManager) SaveStrategy(ctx context.Context, strategy *domain.PricingStrategy) error {
	return m.repo.SavePricingStrategy(ctx, strategy)
}
