package application

import (
	"context"
	"log/slog"
	"time"

	"github.com/wyfcoding/ecommerce/internal/dynamic_pricing/domain"
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

	inventoryFactor := 1.0
	if req.TotalStock > 0 {
		ratio := float64(req.CurrentStock) / float64(req.TotalStock)
		if ratio < 0.2 {
			inventoryFactor = 1.1
		} else if ratio > 0.8 {
			inventoryFactor = 0.9
		}
	}

	demandFactor := 1.0
	if req.AverageDailyDemand > 0 {
		ratio := float64(req.DailyDemand) / float64(req.AverageDailyDemand)
		if ratio > 1.5 {
			demandFactor = 1.1
		} else if ratio < 0.5 {
			demandFactor = 0.9
		}
	}

	competitorFactor := 1.0
	if req.CompetitorPrice > 0 {
		if req.CompetitorPrice < req.BasePrice {
			competitorFactor = 0.95
		}
	}

	adjustment := inventoryFactor * demandFactor * competitorFactor
	finalPrice := int64(float64(req.BasePrice) * adjustment)

	if strategy.MinPrice > 0 && finalPrice < strategy.MinPrice {
		finalPrice = strategy.MinPrice
	}
	if strategy.MaxPrice > 0 && finalPrice > strategy.MaxPrice {
		finalPrice = strategy.MaxPrice
	}

	price := &domain.DynamicPrice{
		SKUID:            req.SKUID,
		BasePrice:        req.BasePrice,
		FinalPrice:       finalPrice,
		PriceAdjustment:  adjustment,
		InventoryFactor:  inventoryFactor,
		DemandFactor:     demandFactor,
		CompetitorFactor: competitorFactor,
		EffectiveTime:    time.Now(),
		ExpiryTime:       time.Now().Add(24 * time.Hour),
	}

	if err := m.repo.SaveDynamicPrice(ctx, price); err != nil {
		m.logger.ErrorContext(ctx, "failed to save dynamic price", "sku_id", req.SKUID, "error", err)
		return nil, err
	}
	m.logger.InfoContext(ctx, "dynamic price calculated successfully", "sku_id", req.SKUID, "final_price", finalPrice)

	return price, nil
}

// SaveStrategy 保存（创建或更新）一个定价策略。
func (m *DynamicPricingManager) SaveStrategy(ctx context.Context, strategy *domain.PricingStrategy) error {
	return m.repo.SavePricingStrategy(ctx, strategy)
}
