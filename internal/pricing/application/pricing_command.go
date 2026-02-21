package application

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/wyfcoding/ecommerce/internal/pricing/domain"
	algorithm "github.com/wyfcoding/pkg/algos/finance"
	"github.com/wyfcoding/pkg/messagequeue"
	"github.com/wyfcoding/pkg/utils"
)

// PricingCommandService 处理定价规则、历史记录及动态定价算法。
type PricingCommandService struct {
	repo      domain.PricingRepository
	publisher messagequeue.EventPublisher
	logger    *slog.Logger
}

// NewPricingCommandService creates a new PricingCommandService instance.
func NewPricingCommandService(repo domain.PricingRepository, publisher messagequeue.EventPublisher, logger *slog.Logger) *PricingCommandService {
	return &PricingCommandService{
		repo:      repo,
		publisher: publisher,
		logger:    logger,
	}
}

// CreateRule 创建或更新一个定价规则。
func (m *PricingCommandService) CreateRule(ctx context.Context, rule *domain.PricingRule) error {
	return m.repo.WithTx(ctx, func(tx any) error {
		if err := m.repo.SaveRuleInTx(ctx, tx, rule); err != nil {
			m.logger.ErrorContext(ctx, "failed to create pricing rule", "rule_id", rule.ID, "error", err)
			return err
		}
		event := &domain.PricingRuleUpdatedEvent{
			RuleID:    rule.ID,
			Name:      rule.Name,
			Type:      string(rule.Strategy),
			Timestamp: time.Now(),
		}
		return m.publisher.PublishInTx(ctx, tx, domain.PricingRuleUpdatedEventType, fmt.Sprintf("%d", rule.ID), event)
	})
}

// CalculateDynamicPrice 计算给定SKU的动态价格 (集成自原 dynamicpricing 服务)。
func (m *PricingCommandService) CalculateDynamicPrice(ctx context.Context, req *domain.PricingRequest) (*domain.DynamicPrice, error) {
	// 1. 获取定价规则 (Strategy)
	rule, err := m.repo.GetActiveRule(ctx, 0, req.SKUID)
	if err != nil || rule == nil {
		rule = &domain.PricingRule{
			Strategy: domain.PricingStrategyDynamic,
			MinPrice: uint64(float64(req.BasePrice) * 0.5),
			MaxPrice: uint64(float64(req.BasePrice) * 2.0),
			Enabled:  true,
		}
	}

	// 2. 获取价格弹性
	elasticityVal := 1.0
	if el, err := m.repo.GetPriceElasticity(ctx, req.SKUID); err == nil && el != nil {
		elasticityVal = el.Elasticity
	}

	// 3. 初始化定价引擎
	engine := algorithm.NewPricingEngine(req.BasePrice, int64(rule.MinPrice), int64(rule.MaxPrice), elasticityVal)
	var finalPrice int64
	var factors domain.DynamicPrice

	// 4. 执行核心算法
	switch rule.Strategy {
	case domain.PricingStrategyProfitMaximization:
		history, _ := m.repo.GetDynamicPriceHistory(ctx, req.SKUID, 30)
		var demandData []algorithm.DemandData
		for _, h := range history {
			demandData = append(demandData, algorithm.DemandData{
				Price:  h.FinalPrice,
				Demand: 100, // 简化处理，实际应从订单或分析服务获取
			})
		}
		demandFunc := func(p int64) int64 { return engine.PredictDemand(p, demandData) }
		cost := int64(float64(req.BasePrice) * 0.7)
		finalPrice = engine.OptimalPriceForProfit(cost, demandFunc)
	case domain.PricingStrategyCompetitive:
		compInfo, _ := m.repo.GetCompetitorPriceInfo(ctx, req.SKUID)
		var compPrices []int64
		if compInfo != nil {
			for _, c := range compInfo.Competitors {
				compPrices = append(compPrices, c.Price)
			}
		}
		finalPrice = engine.CompetitivePricing(compPrices, "average")
	default:
		now := time.Now()
		demandLevel := 0.5
		if req.AverageDailyDemand > 0 {
			demandLevel = 0.5 * float64(req.DailyDemand) / float64(req.AverageDailyDemand)
		}

		factors_alg := algorithm.PricingFactors{
			Stock:           req.CurrentStock,
			TotalStock:      req.TotalStock,
			DemandLevel:     demandLevel,
			CompetitorPrice: req.CompetitorPrice,
			TimeOfDay:       now.Hour(),
			DayOfWeek:       int(now.Weekday()),
			IsHoliday:       utils.IsHoliday(now),
			UserLevel:       5, // Default
			SeasonFactor:    0.5,
		}
		result := engine.CalculatePrice(factors_alg)
		finalPrice = result.FinalPrice
		factors.InventoryFactor = result.InventoryFactor
		factors.DemandFactor = result.DemandFactor
		factors.CompetitorFactor = result.CompetitorFactor
	}

	// 5. 保存结果并发布事件
	adjustment := 1.0
	if req.BasePrice > 0 {
		adjustment = float64(finalPrice) / float64(req.BasePrice)
	}

	price := &domain.DynamicPrice{
		SKUID:            req.SKUID,
		BasePrice:        req.BasePrice,
		FinalPrice:       finalPrice,
		PriceAdjustment:  adjustment,
		InventoryFactor:  factors.InventoryFactor,
		DemandFactor:     factors.DemandFactor,
		CompetitorFactor: factors.CompetitorFactor,
		EffectiveTime:    time.Now(),
		ExpiryTime:       time.Now().Add(24 * time.Hour),
	}

	err = m.repo.WithTx(ctx, func(tx any) error {
		if err := m.repo.SaveDynamicPriceInTx(ctx, tx, price); err != nil {
			return err
		}
		return m.publisher.PublishInTx(ctx, tx, "DynamicPriceCalculated", fmt.Sprintf("%d", req.SKUID), price)
	})

	return price, err
}

// UpdateCompetitorPrice 更新竞品价格信息。
func (m *PricingCommandService) UpdateCompetitorPrice(ctx context.Context, skuID uint64, competitorName string, price int64, url string) error {
	info, err := m.repo.GetCompetitorPriceInfo(ctx, skuID)
	if err != nil {
		return err
	}
	if info == nil {
		info = &domain.CompetitorPriceInfo{SKUID: skuID, LastUpdated: time.Now()}
	}

	return m.repo.WithTx(ctx, func(tx any) error {
		if info.ID == 0 {
			if err := m.repo.SaveCompetitorPriceInfoInTx(ctx, tx, info); err != nil {
				return err
			}
		}
		sub := &domain.CompetitorPrice{
			InfoID:         uint64(info.ID),
			CompetitorName: competitorName,
			Price:          price,
			URL:            url,
			LastUpdated:    time.Now(),
		}
		return m.repo.SaveCompetitorPriceInTx(ctx, tx, sub)
	})
}

// RecordHistory 记录价格变动历史。
func (m *PricingCommandService) RecordHistory(ctx context.Context, productID, skuID, price, oldPrice uint64, reason string) error {
	var changeRate float64
	if oldPrice > 0 {
		changeRate = float64(price-oldPrice) / float64(oldPrice) * 100
	}

	history := &domain.PriceHistory{
		ProductID:  productID,
		SkuID:      skuID,
		Price:      price,
		OldPrice:   oldPrice,
		ChangeRate: changeRate,
		Reason:     reason,
	}

	return m.repo.WithTx(ctx, func(tx any) error {
		if err := m.repo.SaveHistoryInTx(ctx, tx, history); err != nil {
			return err
		}
		return m.publisher.PublishInTx(ctx, tx, domain.PriceHistoryRecordedEventType, fmt.Sprintf("%d", history.ID), history)
	})
}
