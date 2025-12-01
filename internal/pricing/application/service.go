package application

import (
	"context"

	"github.com/wyfcoding/ecommerce/internal/pricing/domain/entity"
	"github.com/wyfcoding/ecommerce/internal/pricing/domain/repository"

	"log/slog"
)

type PricingService struct {
	repo   repository.PricingRepository
	logger *slog.Logger
}

func NewPricingService(repo repository.PricingRepository, logger *slog.Logger) *PricingService {
	return &PricingService{
		repo:   repo,
		logger: logger,
	}
}

// CreateRule 创建规则
func (s *PricingService) CreateRule(ctx context.Context, rule *entity.PricingRule) error {
	if err := s.repo.SaveRule(ctx, rule); err != nil {
		s.logger.ErrorContext(ctx, "failed to create pricing rule", "rule_id", rule.ID, "error", err)
		return err
	}
	s.logger.InfoContext(ctx, "pricing rule created successfully", "rule_id", rule.ID)
	return nil
}

// CalculatePrice 计算价格
func (s *PricingService) CalculatePrice(ctx context.Context, productID, skuID uint64, demand, competition float64) (uint64, error) {
	rule, err := s.repo.GetActiveRule(ctx, productID, skuID)
	if err != nil {
		return 0, err
	}
	if rule == nil {
		// No active rule, return 0 or default price (here 0 indicates no price calculated)
		return 0, nil
	}

	price := rule.CalculatePrice(demand, competition)
	return price, nil
}

// RecordHistory 记录历史
func (s *PricingService) RecordHistory(ctx context.Context, productID, skuID, price, oldPrice uint64, reason string) error {
	var changeRate float64
	if oldPrice > 0 {
		changeRate = float64(price-oldPrice) / float64(oldPrice) * 100
	}

	history := &entity.PriceHistory{
		ProductID:  productID,
		SkuID:      skuID,
		Price:      price,
		OldPrice:   oldPrice,
		ChangeRate: changeRate,
		Reason:     reason,
	}
	if err := s.repo.SaveHistory(ctx, history); err != nil {
		s.logger.ErrorContext(ctx, "failed to record price history", "product_id", productID, "sku_id", skuID, "error", err)
		return err
	}
	s.logger.InfoContext(ctx, "price history recorded successfully", "product_id", productID, "sku_id", skuID, "price", price)
	return nil
}

// ListRules 规则列表
func (s *PricingService) ListRules(ctx context.Context, productID uint64, page, pageSize int) ([]*entity.PricingRule, int64, error) {
	offset := (page - 1) * pageSize
	return s.repo.ListRules(ctx, productID, offset, pageSize)
}

// ListHistory 历史列表
func (s *PricingService) ListHistory(ctx context.Context, productID, skuID uint64, page, pageSize int) ([]*entity.PriceHistory, int64, error) {
	offset := (page - 1) * pageSize
	return s.repo.ListHistory(ctx, productID, skuID, offset, pageSize)
}
