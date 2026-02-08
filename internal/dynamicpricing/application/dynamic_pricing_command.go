package application

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/wyfcoding/ecommerce/internal/dynamicpricing/domain"
	algorithm "github.com/wyfcoding/pkg/algorithm/finance"
	"github.com/wyfcoding/pkg/messagequeue"
	"github.com/wyfcoding/pkg/utils"
)

// DynamicPricingCommandService 处理动态定价的写操作。
type DynamicPricingCommandService struct {
	repo      domain.PricingRepository
	publisher messagequeue.EventPublisher
	logger    *slog.Logger
}

// NewDynamicPricingCommandService 创建并返回一个新的 DynamicPricingCommandService 实例。
func NewDynamicPricingCommandService(repo domain.PricingRepository, publisher messagequeue.EventPublisher, logger *slog.Logger) *DynamicPricingCommandService {
	return &DynamicPricingCommandService{
		repo:      repo,
		publisher: publisher,
		logger:    logger,
	}
}

// CalculatePrice 计算给定SKU的动态价格。
func (m *DynamicPricingCommandService) CalculatePrice(ctx context.Context, req *domain.PricingRequest) (*domain.DynamicPrice, error) {
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
	var algoFactors domain.DynamicPrice

	// 4. 根据策略类型执行计算
	switch strategy.StrategyType {
	case "profit_maximization":
		// 利润最大化策略
		history, _ := m.repo.GetPriceHistory(ctx, req.SKUID, 30)
		var demandData []algorithm.DemandData
		for _, h := range history {
			demandData = append(demandData, algorithm.DemandData{
				Price:  h.Price,
				Demand: int64(h.Quantity),
			})
		}

		demandFunc := func(p int64) int64 {
			return engine.PredictDemand(p, demandData)
		}

		cost := int64(float64(req.BasePrice) * 0.7)
		finalPrice = engine.OptimalPriceForProfit(cost, demandFunc)

		algoFactors.DemandFactor = 1.0
		algoFactors.InventoryFactor = 1.0
	case "competitive":
		compInfo, _ := m.repo.GetCompetitorPriceInfo(ctx, req.SKUID)
		var compPrices []int64
		if compInfo != nil {
			for _, c := range compInfo.Competitors {
				compPrices = append(compPrices, c.Price)
			}
		}
		finalPrice = engine.CompetitivePricing(compPrices, "average")
		algoFactors.CompetitorFactor = float64(finalPrice) / float64(req.BasePrice)
	default:
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

		compPrice := req.CompetitorPrice
		if compPrice == 0 {
			if compInfo, err := m.repo.GetCompetitorPriceInfo(ctx, req.SKUID); err == nil && compInfo != nil {
				compPrice = compInfo.LowestPrice
			}
		}

		factors := algorithm.PricingFactors{
			Stock:           req.CurrentStock,
			TotalStock:      req.TotalStock,
			DemandLevel:     demandLevel,
			CompetitorPrice: compPrice,
			TimeOfDay:       now.Hour(),
			DayOfWeek:       int(now.Weekday()),
			IsHoliday:       utils.IsHoliday(now),
			UserLevel:       userLevel,
			SeasonFactor:    0.5,
		}

		result := engine.CalculatePrice(factors)
		finalPrice = result.FinalPrice
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

	if err := m.repo.WithTx(ctx, func(tx any) error {
		if err := m.repo.SaveDynamicPriceInTx(ctx, tx, price); err != nil {
			m.logger.ErrorContext(ctx, "failed to save dynamic price", "sku_id", req.SKUID, "error", err)
			return err
		}
		if m.publisher == nil {
			return nil
		}
		return m.publisher.PublishInTx(ctx, tx, domain.DynamicPriceCalculatedEventType, fmt.Sprintf("%d", req.SKUID), &domain.PriceCalculatedEvent{
			SKUID:      req.SKUID,
			BasePrice:  req.BasePrice,
			FinalPrice: finalPrice,
			Timestamp:  time.Now(),
		})
	}); err != nil {
		return nil, err
	}

	m.logger.InfoContext(ctx, "dynamic price calculated successfully", "sku_id", req.SKUID, "strategy", strategy.StrategyType, "final_price", finalPrice)
	return price, nil
}

// SaveStrategy 保存（创建或更新）一个定价策略。
func (m *DynamicPricingCommandService) SaveStrategy(ctx context.Context, strategy *domain.PricingStrategy) error {
	if strategy == nil {
		return nil
	}

	existing, _ := m.repo.GetPricingStrategy(ctx, strategy.SKUID)
	created := existing == nil

	if err := m.repo.WithTx(ctx, func(tx any) error {
		if err := m.repo.SavePricingStrategyInTx(ctx, tx, strategy); err != nil {
			return err
		}
		if m.publisher == nil {
			return nil
		}
		event := &domain.StrategyUpdatedEvent{
			StrategyID:   uint64(strategy.ID),
			SKUID:        strategy.SKUID,
			StrategyType: strategy.StrategyType,
			Enabled:      strategy.Enabled,
			Timestamp:    time.Now(),
		}
		topic := domain.PricingStrategyUpdatedEventType
		if created {
			topic = domain.PricingStrategyCreatedEventType
		}
		return m.publisher.PublishInTx(ctx, tx, topic, fmt.Sprintf("%d", strategy.SKUID), event)
	}); err != nil {
		return err
	}

	return nil
}

// UpdateCompetitorPrice 记录并更新竞争对手对特定 SKU 的最新报价。
func (m *DynamicPricingCommandService) UpdateCompetitorPrice(ctx context.Context, skuID uint64, competitorName string, price int64, url string) error {
	// 1. 获取（或创建）汇总信息
	info, err := m.repo.GetCompetitorPriceInfo(ctx, skuID)
	if err != nil {
		return fmt.Errorf("failed to get competitor price info: %w", err)
	}

	if info == nil {
		info = &domain.CompetitorPriceInfo{
			SKUID:       skuID,
			LastUpdated: time.Now(),
		}
	}

	return m.repo.WithTx(ctx, func(tx any) error {
		// 2. 如果汇总信息不存在，先创建
		if info.ID == 0 {
			if err := m.repo.SaveCompetitorPriceInfoInTx(ctx, tx, info); err != nil {
				return err
			}
		}

		// 3. 保存明细数据
		newSub := &domain.CompetitorPrice{
			InfoID:         uint64(info.ID),
			CompetitorName: competitorName,
			Price:          price,
			URL:            url,
			LastUpdated:    time.Now(),
		}

		if err := m.repo.SaveCompetitorPriceInTx(ctx, tx, newSub); err != nil {
			return err
		}

		// 4. (可选) 可以在此处触发重新计算汇总信息（平均价、最低价等）
		return nil
	})
}
