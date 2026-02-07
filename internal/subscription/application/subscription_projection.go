// 生成摘要：新增订阅读模型投影服务，消费事件后刷新 Redis/ES 读侧。
package application

import (
	"context"
	"log/slog"

	"github.com/wyfcoding/ecommerce/internal/subscription/domain"
)

// SubscriptionProjectionService 负责将订阅事件投影到读模型。
type SubscriptionProjectionService struct {
	repo                   domain.SubscriptionRepository
	planReadRepo           domain.SubscriptionPlanReadRepository
	subscriptionReadRepo   domain.SubscriptionReadRepository
	subscriptionSearchRepo domain.SubscriptionSearchRepository
	logger                 *slog.Logger
}

// NewSubscriptionProjectionService 创建投影服务。
func NewSubscriptionProjectionService(
	repo domain.SubscriptionRepository,
	planReadRepo domain.SubscriptionPlanReadRepository,
	subscriptionReadRepo domain.SubscriptionReadRepository,
	subscriptionSearchRepo domain.SubscriptionSearchRepository,
	logger *slog.Logger,
) *SubscriptionProjectionService {
	return &SubscriptionProjectionService{
		repo:                   repo,
		planReadRepo:           planReadRepo,
		subscriptionReadRepo:   subscriptionReadRepo,
		subscriptionSearchRepo: subscriptionSearchRepo,
		logger:                 logger,
	}
}

func (s *SubscriptionProjectionService) OnPlanCreated(ctx context.Context, event *domain.SubscriptionPlanCreatedEvent) error {
	if event == nil {
		return nil
	}
	return s.refreshPlan(ctx, event.PlanID)
}

func (s *SubscriptionProjectionService) OnPlanUpdated(ctx context.Context, event *domain.SubscriptionPlanUpdatedEvent) error {
	if event == nil {
		return nil
	}
	return s.refreshPlan(ctx, event.PlanID)
}

func (s *SubscriptionProjectionService) OnPlanDeleted(ctx context.Context, event *domain.SubscriptionPlanDeletedEvent) error {
	if event == nil {
		return nil
	}
	if s.planReadRepo != nil {
		_ = s.planReadRepo.Delete(ctx, event.PlanID)
	}
	return nil
}

func (s *SubscriptionProjectionService) OnSubscriptionCreated(ctx context.Context, event *domain.SubscriptionCreatedEvent) error {
	if event == nil {
		return nil
	}
	return s.refreshSubscription(ctx, event.SubscriptionID)
}

func (s *SubscriptionProjectionService) OnSubscriptionCancelled(ctx context.Context, event *domain.SubscriptionCancelledEvent) error {
	if event == nil {
		return nil
	}
	return s.refreshSubscription(ctx, event.SubscriptionID)
}

func (s *SubscriptionProjectionService) OnSubscriptionRenewed(ctx context.Context, event *domain.SubscriptionRenewedEvent) error {
	if event == nil {
		return nil
	}
	return s.refreshSubscription(ctx, event.SubscriptionID)
}

func (s *SubscriptionProjectionService) refreshPlan(ctx context.Context, planID uint64) error {
	if s.planReadRepo == nil {
		return nil
	}
	plan, err := s.repo.GetPlan(ctx, planID)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to load plan for projection", "plan_id", planID, "error", err)
		return err
	}
	if plan == nil {
		_ = s.planReadRepo.Delete(ctx, planID)
		return nil
	}
	if err := s.planReadRepo.Save(ctx, plan); err != nil {
		s.logger.ErrorContext(ctx, "failed to save plan cache", "plan_id", planID, "error", err)
		return err
	}
	return nil
}

func (s *SubscriptionProjectionService) refreshSubscription(ctx context.Context, subscriptionID uint64) error {
	if s.subscriptionReadRepo == nil && s.subscriptionSearchRepo == nil {
		return nil
	}
	sub, err := s.repo.GetSubscription(ctx, subscriptionID)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to load subscription for projection", "subscription_id", subscriptionID, "error", err)
		return err
	}
	if sub == nil {
		if s.subscriptionReadRepo != nil {
			_ = s.subscriptionReadRepo.Delete(ctx, subscriptionID)
		}
		if s.subscriptionSearchRepo != nil {
			_ = s.subscriptionSearchRepo.Delete(ctx, subscriptionID)
		}
		return nil
	}
	if s.subscriptionReadRepo != nil {
		if err := s.subscriptionReadRepo.Save(ctx, sub); err != nil {
			s.logger.ErrorContext(ctx, "failed to save subscription cache", "subscription_id", subscriptionID, "error", err)
			return err
		}
	}
	if s.subscriptionSearchRepo != nil {
		if err := s.subscriptionSearchRepo.Index(ctx, sub); err != nil {
			s.logger.ErrorContext(ctx, "failed to index subscription", "subscription_id", subscriptionID, "error", err)
			return err
		}
	}
	return nil
}
