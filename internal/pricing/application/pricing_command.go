package application

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/wyfcoding/ecommerce/internal/pricing/domain"
)

// PricingCommandService 处理定价规则和历史记录的写操作。
type PricingCommandService struct {
	repo      domain.PricingRepository
	publisher domain.EventPublisher
	logger    *slog.Logger
}

// NewPricingCommandService creates a new PricingCommandService instance.
func NewPricingCommandService(repo domain.PricingRepository, publisher domain.EventPublisher, logger *slog.Logger) *PricingCommandService {
	return &PricingCommandService{
		repo:      repo,
		publisher: publisher,
		logger:    logger,
	}
}

// CreateRule 创建一个新的定价规则。
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
		if err := m.publisher.PublishInTx(ctx, tx, domain.PricingRuleUpdatedEventType, fmt.Sprintf("%d", rule.ID), event); err != nil {
			m.logger.ErrorContext(ctx, "failed to publish pricing rule event", "rule_id", rule.ID, "error", err)
			return err
		}
		m.logger.InfoContext(ctx, "pricing rule created successfully", "rule_id", rule.ID)
		return nil
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
			m.logger.ErrorContext(ctx, "failed to record price history", "product_id", productID, "sku_id", skuID, "error", err)
			return err
		}
		event := &domain.PriceHistoryRecordedEvent{
			HistoryID: history.ID,
			ProductID: history.ProductID,
			SKUID:     history.SkuID,
			Price:     history.Price,
			OldPrice:  history.OldPrice,
			Timestamp: time.Now(),
		}
		if err := m.publisher.PublishInTx(ctx, tx, domain.PriceHistoryRecordedEventType, fmt.Sprintf("%d", history.ID), event); err != nil {
			m.logger.ErrorContext(ctx, "failed to publish price history event", "history_id", history.ID, "error", err)
			return err
		}
		m.logger.InfoContext(ctx, "price history recorded successfully", "product_id", productID, "sku_id", skuID, "price", price)
		return nil
	})
}
