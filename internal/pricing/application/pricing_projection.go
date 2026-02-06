// 生成摘要：新增定价读模型投影服务，消费事件后刷新 Redis/ES 读侧。
package application

import (
	"context"
	"log/slog"

	"github.com/wyfcoding/ecommerce/internal/pricing/domain"
)

// PricingProjectionService 负责将定价事件投影到读模型。
type PricingProjectionService struct {
	repo              domain.PricingRepository
	ruleReadRepo      domain.PricingRuleReadRepository
	historySearchRepo domain.PriceHistorySearchRepository
	logger            *slog.Logger
}

// NewPricingProjectionService 创建定价投影服务。
func NewPricingProjectionService(
	repo domain.PricingRepository,
	ruleReadRepo domain.PricingRuleReadRepository,
	historySearchRepo domain.PriceHistorySearchRepository,
	logger *slog.Logger,
) *PricingProjectionService {
	return &PricingProjectionService{
		repo:              repo,
		ruleReadRepo:      ruleReadRepo,
		historySearchRepo: historySearchRepo,
		logger:            logger,
	}
}

func (s *PricingProjectionService) OnPricingRuleUpdated(ctx context.Context, event *domain.PricingRuleUpdatedEvent) error {
	if event == nil {
		return nil
	}
	if s.ruleReadRepo == nil {
		return nil
	}
	rule, err := s.repo.GetRule(ctx, event.RuleID)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to load rule for projection", "rule_id", event.RuleID, "error", err)
		return err
	}
	if rule == nil {
		_ = s.ruleReadRepo.Delete(ctx, event.RuleID)
		return nil
	}
	if err := s.ruleReadRepo.Save(ctx, rule); err != nil {
		s.logger.ErrorContext(ctx, "failed to save rule read model", "rule_id", event.RuleID, "error", err)
		return err
	}
	return nil
}

func (s *PricingProjectionService) OnPriceHistoryRecorded(ctx context.Context, event *domain.PriceHistoryRecordedEvent) error {
	if event == nil || s.historySearchRepo == nil {
		return nil
	}
	history, err := s.repo.GetHistory(ctx, event.HistoryID)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to load history for projection", "history_id", event.HistoryID, "error", err)
		return err
	}
	if history == nil {
		_ = s.historySearchRepo.Delete(ctx, event.HistoryID)
		return nil
	}
	if err := s.historySearchRepo.Index(ctx, history); err != nil {
		s.logger.ErrorContext(ctx, "failed to index price history", "history_id", event.HistoryID, "error", err)
		return err
	}
	return nil
}
