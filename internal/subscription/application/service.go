package application

import (
	"context"
	"github.com/wyfcoding/ecommerce/internal/subscription/domain/entity"
	"github.com/wyfcoding/ecommerce/internal/subscription/domain/repository"
	"errors"
	"time"

	"log/slog"
)

type SubscriptionService struct {
	repo   repository.SubscriptionRepository
	logger *slog.Logger
}

func NewSubscriptionService(repo repository.SubscriptionRepository, logger *slog.Logger) *SubscriptionService {
	return &SubscriptionService{
		repo:   repo,
		logger: logger,
	}
}

// CreatePlan 创建计划
func (s *SubscriptionService) CreatePlan(ctx context.Context, name, desc string, price uint64, duration int32, features []string) (*entity.SubscriptionPlan, error) {
	plan := &entity.SubscriptionPlan{
		Name:        name,
		Description: desc,
		Price:       price,
		Duration:    duration,
		Features:    features,
		Enabled:     true,
	}
	if err := s.repo.SavePlan(ctx, plan); err != nil {
		return nil, err
	}
	return plan, nil
}

// Subscribe 订阅
func (s *SubscriptionService) Subscribe(ctx context.Context, userID, planID uint64) (*entity.Subscription, error) {
	// Check if active subscription exists
	active, err := s.repo.GetActiveSubscription(ctx, userID)
	if err != nil {
		return nil, err
	}
	if active != nil {
		return nil, errors.New("user already has an active subscription")
	}

	// Get Plan
	plan, err := s.repo.GetPlan(ctx, planID)
	if err != nil {
		return nil, err
	}
	if plan == nil || !plan.Enabled {
		return nil, errors.New("plan not found or disabled")
	}

	now := time.Now()
	sub := &entity.Subscription{
		UserID:    userID,
		PlanID:    planID,
		Status:    entity.SubscriptionStatusActive,
		StartDate: now,
		EndDate:   now.AddDate(0, 0, int(plan.Duration)),
		AutoRenew: true,
	}

	if err := s.repo.SaveSubscription(ctx, sub); err != nil {
		return nil, err
	}
	return sub, nil
}

// Cancel 取消订阅
func (s *SubscriptionService) Cancel(ctx context.Context, id uint64) error {
	sub, err := s.repo.GetSubscription(ctx, id)
	if err != nil {
		return err
	}
	if sub == nil {
		return errors.New("subscription not found")
	}

	sub.Status = entity.SubscriptionStatusCanceled
	sub.AutoRenew = false
	now := time.Now()
	sub.CanceledAt = &now

	return s.repo.SaveSubscription(ctx, sub)
}

// Renew 续订 (Manual or Auto)
func (s *SubscriptionService) Renew(ctx context.Context, id uint64) error {
	sub, err := s.repo.GetSubscription(ctx, id)
	if err != nil {
		return err
	}
	if sub == nil {
		return errors.New("subscription not found")
	}

	plan, err := s.repo.GetPlan(ctx, sub.PlanID)
	if err != nil {
		return err
	}

	// Extend end date
	if sub.EndDate.Before(time.Now()) {
		sub.EndDate = time.Now().AddDate(0, 0, int(plan.Duration))
	} else {
		sub.EndDate = sub.EndDate.AddDate(0, 0, int(plan.Duration))
	}
	sub.Status = entity.SubscriptionStatusActive

	return s.repo.SaveSubscription(ctx, sub)
}

// ListPlans 获取计划列表
func (s *SubscriptionService) ListPlans(ctx context.Context) ([]*entity.SubscriptionPlan, error) {
	return s.repo.ListPlans(ctx, true)
}

// ListSubscriptions 获取订阅列表
func (s *SubscriptionService) ListSubscriptions(ctx context.Context, userID uint64, page, pageSize int) ([]*entity.Subscription, int64, error) {
	offset := (page - 1) * pageSize
	return s.repo.ListSubscriptions(ctx, userID, nil, offset, pageSize)
}
