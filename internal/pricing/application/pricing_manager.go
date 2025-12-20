package application

import (
	"context"
	"log/slog"

	"github.com/wyfcoding/ecommerce/internal/pricing/domain"
)

// PricingManager 处理定价规则和历史记录的写操作。
type PricingManager struct {
	repo   domain.PricingRepository
	logger *slog.Logger
}

// NewPricingManager creates a new PricingManager instance.
func NewPricingManager(repo domain.PricingRepository, logger *slog.Logger) *PricingManager {
	return &PricingManager{
		repo:   repo,
		logger: logger,
	}
}

// CreateRule 创建一个新的定价规则。
func (m *PricingManager) CreateRule(ctx context.Context, rule *domain.PricingRule) error {
	if err := m.repo.SaveRule(ctx, rule); err != nil {
		m.logger.ErrorContext(ctx, "failed to create pricing rule", "rule_id", rule.ID, "error", err)
		return err
	}
	m.logger.InfoContext(ctx, "pricing rule created successfully", "rule_id", rule.ID)
	return nil
}

// RecordHistory 记录价格变动历史。
func (m *PricingManager) RecordHistory(ctx context.Context, productID, skuID, price, oldPrice uint64, reason string) error {
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

	if err := m.repo.SaveHistory(ctx, history); err != nil {
		m.logger.ErrorContext(ctx, "failed to record price history", "product_id", productID, "sku_id", skuID, "error", err)
		return err
	}
	m.logger.InfoContext(ctx, "price history recorded successfully", "product_id", productID, "sku_id", skuID, "price", price)
	return nil
}
