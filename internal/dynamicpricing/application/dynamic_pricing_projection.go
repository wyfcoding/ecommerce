package application

import (
	"context"
	"log/slog"

	"github.com/wyfcoding/ecommerce/internal/dynamicpricing/domain"
)

// DynamicPricingProjectionService 负责将定价事件投影到读模型。
type DynamicPricingProjectionService struct {
	repo           domain.PricingRepository
	priceRead      domain.DynamicPriceReadRepository
	strategyRead   domain.PricingStrategyReadRepository
	strategySearch domain.PricingStrategySearchRepository
	logger         *slog.Logger
}

// NewDynamicPricingProjectionService 创建投影服务。
func NewDynamicPricingProjectionService(
	repo domain.PricingRepository,
	priceRead domain.DynamicPriceReadRepository,
	strategyRead domain.PricingStrategyReadRepository,
	strategySearch domain.PricingStrategySearchRepository,
	logger *slog.Logger,
) *DynamicPricingProjectionService {
	return &DynamicPricingProjectionService{
		repo:           repo,
		priceRead:      priceRead,
		strategyRead:   strategyRead,
		strategySearch: strategySearch,
		logger:         logger,
	}
}

func (s *DynamicPricingProjectionService) OnPriceCalculated(ctx context.Context, event *domain.PriceCalculatedEvent) error {
	if event == nil {
		return nil
	}
	return s.refreshLatestPrice(ctx, event.SKUID)
}

func (s *DynamicPricingProjectionService) OnStrategyChanged(ctx context.Context, event *domain.StrategyUpdatedEvent) error {
	if event == nil {
		return nil
	}
	return s.refreshStrategy(ctx, event.SKUID)
}

func (s *DynamicPricingProjectionService) refreshLatestPrice(ctx context.Context, skuID uint64) error {
	if s.priceRead == nil {
		return nil
	}
	price, err := s.repo.GetLatestDynamicPrice(ctx, skuID)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to load latest price for projection", "sku_id", skuID, "error", err)
		return err
	}
	if price == nil {
		_ = s.priceRead.DeleteLatest(ctx, skuID)
		return nil
	}
	if err := s.priceRead.SaveLatest(ctx, price); err != nil {
		s.logger.ErrorContext(ctx, "failed to save latest price cache", "sku_id", skuID, "error", err)
		return err
	}
	return nil
}

func (s *DynamicPricingProjectionService) refreshStrategy(ctx context.Context, skuID uint64) error {
	strategy, err := s.repo.GetPricingStrategy(ctx, skuID)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to load strategy for projection", "sku_id", skuID, "error", err)
		return err
	}
	if strategy == nil {
		if s.strategyRead != nil {
			_ = s.strategyRead.Delete(ctx, skuID)
		}
		if s.strategySearch != nil {
			_ = s.strategySearch.Delete(ctx, skuID)
		}
		return nil
	}
	if s.strategyRead != nil {
		if err := s.strategyRead.Save(ctx, strategy); err != nil {
			s.logger.ErrorContext(ctx, "failed to save strategy cache", "sku_id", skuID, "error", err)
			return err
		}
	}
	if s.strategySearch != nil {
		if err := s.strategySearch.Index(ctx, strategy); err != nil {
			s.logger.ErrorContext(ctx, "failed to index strategy", "sku_id", skuID, "error", err)
			return err
		}
	}
	return nil
}
