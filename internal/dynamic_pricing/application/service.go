package application

import (
	"context"
	"time"

	"github.com/wyfcoding/ecommerce/internal/dynamic_pricing/domain/entity"
	"github.com/wyfcoding/ecommerce/internal/dynamic_pricing/domain/repository"

	"log/slog"
)

type DynamicPricingService struct {
	repo   repository.PricingRepository
	logger *slog.Logger
}

func NewDynamicPricingService(repo repository.PricingRepository, logger *slog.Logger) *DynamicPricingService {
	return &DynamicPricingService{
		repo:   repo,
		logger: logger,
	}
}

// CalculatePrice 计算动态价格
func (s *DynamicPricingService) CalculatePrice(ctx context.Context, req *entity.PricingRequest) (*entity.DynamicPrice, error) {
	// 1. Get Strategy
	strategy, err := s.repo.GetPricingStrategy(ctx, req.SKUID)
	if err != nil {
		// Default strategy if not found
		strategy = &entity.PricingStrategy{
			StrategyType: "fixed",
			MinPrice:     req.BasePrice,
			MaxPrice:     req.BasePrice,
		}
	}

	// 2. Calculate factors (simplified logic)
	inventoryFactor := 1.0
	if req.TotalStock > 0 {
		ratio := float64(req.CurrentStock) / float64(req.TotalStock)
		if ratio < 0.2 {
			inventoryFactor = 1.1 // Low stock, increase price
		} else if ratio > 0.8 {
			inventoryFactor = 0.9 // High stock, decrease price
		}
	}

	demandFactor := 1.0
	if req.AverageDailyDemand > 0 {
		ratio := float64(req.DailyDemand) / float64(req.AverageDailyDemand)
		if ratio > 1.5 {
			demandFactor = 1.1 // High demand
		} else if ratio < 0.5 {
			demandFactor = 0.9 // Low demand
		}
	}

	competitorFactor := 1.0
	if req.CompetitorPrice > 0 {
		if req.CompetitorPrice < req.BasePrice {
			competitorFactor = 0.95 // Match competitor slightly
		}
	}

	// 3. Calculate Final Price
	adjustment := inventoryFactor * demandFactor * competitorFactor
	finalPrice := int64(float64(req.BasePrice) * adjustment)

	// Apply constraints
	if strategy.MinPrice > 0 && finalPrice < strategy.MinPrice {
		finalPrice = strategy.MinPrice
	}
	if strategy.MaxPrice > 0 && finalPrice > strategy.MaxPrice {
		finalPrice = strategy.MaxPrice
	}

	price := &entity.DynamicPrice{
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

	if err := s.repo.SaveDynamicPrice(ctx, price); err != nil {
		s.logger.ErrorContext(ctx, "failed to save dynamic price", "sku_id", req.SKUID, "error", err)
		return nil, err
	}
	s.logger.InfoContext(ctx, "dynamic price calculated successfully", "sku_id", req.SKUID, "final_price", finalPrice)

	return price, nil
}

// GetLatestPrice 获取最新价格
func (s *DynamicPricingService) GetLatestPrice(ctx context.Context, skuID uint64) (*entity.DynamicPrice, error) {
	return s.repo.GetLatestDynamicPrice(ctx, skuID)
}

// CreateStrategy 创建/更新策略
func (s *DynamicPricingService) SaveStrategy(ctx context.Context, strategy *entity.PricingStrategy) error {
	return s.repo.SavePricingStrategy(ctx, strategy)
}

// ListStrategies 获取策略列表
func (s *DynamicPricingService) ListStrategies(ctx context.Context, page, pageSize int) ([]*entity.PricingStrategy, int64, error) {
	offset := (page - 1) * pageSize
	return s.repo.ListPricingStrategies(ctx, offset, pageSize)
}
