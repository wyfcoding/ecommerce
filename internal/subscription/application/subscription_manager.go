package application

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/wyfcoding/ecommerce/internal/subscription/domain"
)

// SubscriptionManager 处理订阅的写操作。
type SubscriptionManager struct {
	repo   domain.SubscriptionRepository
	logger *slog.Logger
}

// NewSubscriptionManager creates a new SubscriptionManager instance.
func NewSubscriptionManager(repo domain.SubscriptionRepository, logger *slog.Logger) *SubscriptionManager {
	return &SubscriptionManager{
		repo:   repo,
		logger: logger,
	}
}

// CreatePlan creates a new subscription plan.
func (m *SubscriptionManager) CreatePlan(ctx context.Context, name, desc string, price uint64, duration int32, features []string) (*domain.SubscriptionPlan, error) {
	plan := &domain.SubscriptionPlan{
		Name:        name,
		Description: desc,
		Price:       price,
		Duration:    duration,
		Features:    features,
		Enabled:     true,
	}
	if err := m.repo.SavePlan(ctx, plan); err != nil {
		m.logger.ErrorContext(ctx, "failed to create plan", "name", name, "error", err)
		return nil, err
	}
	m.logger.InfoContext(ctx, "plan created successfully", "plan_id", plan.ID, "name", name)
	return plan, nil
}

// Subscribe 为用户订阅计划。
func (m *SubscriptionManager) Subscribe(ctx context.Context, userID, planID uint64) (*domain.Subscription, error) {
	active, err := m.repo.GetActiveSubscription(ctx, userID)
	if err != nil {
		m.logger.ErrorContext(ctx, "failed to check active subscription", "user_id", userID, "error", err)
		return nil, err
	}
	if active != nil {
		return nil, errors.New("user already has an active subscription")
	}

	plan, err := m.repo.GetPlan(ctx, planID)
	if err != nil {
		m.logger.ErrorContext(ctx, "failed to get plan", "plan_id", planID, "error", err)
		return nil, err
	}
	if plan == nil || !plan.Enabled {
		return nil, errors.New("plan not found or disabled")
	}

	now := time.Now()
	sub := &domain.Subscription{
		UserID:    userID,
		PlanID:    planID,
		Status:    domain.SubscriptionStatusActive,
		StartDate: now,
		EndDate:   now.AddDate(0, 0, int(plan.Duration)),
		AutoRenew: true,
	}

	if err := m.repo.SaveSubscription(ctx, sub); err != nil {
		m.logger.ErrorContext(ctx, "failed to save subscription", "user_id", userID, "plan_id", planID, "error", err)
		return nil, err
	}
	m.logger.InfoContext(ctx, "subscription created successfully", "subscription_id", sub.ID, "user_id", userID)
	return sub, nil
}

// Cancel cancels a subscription.
func (m *SubscriptionManager) Cancel(ctx context.Context, id uint64) error {
	sub, err := m.repo.GetSubscription(ctx, id)
	if err != nil {
		return err
	}
	if sub == nil {
		return errors.New("subscription not found")
	}

	sub.Status = domain.SubscriptionStatusCanceled
	sub.AutoRenew = false
	now := time.Now()
	sub.CanceledAt = &now

	if err := m.repo.SaveSubscription(ctx, sub); err != nil {
		m.logger.ErrorContext(ctx, "failed to cancel subscription", "subscription_id", id, "error", err)
		return err
	}
	m.logger.InfoContext(ctx, "subscription canceled successfully", "subscription_id", id)
	return nil
}

// Renew renews a subscription.
func (m *SubscriptionManager) Renew(ctx context.Context, id uint64) error {
	sub, err := m.repo.GetSubscription(ctx, id)
	if err != nil {
		return err
	}
	if sub == nil {
		return errors.New("subscription not found")
	}

	plan, err := m.repo.GetPlan(ctx, sub.PlanID)
	if err != nil {
		m.logger.ErrorContext(ctx, "failed to get plan for renewal", "plan_id", sub.PlanID, "error", err)
		return err
	}
	if plan == nil {
		return errors.New("plan not found for renewal")
	}

	if sub.EndDate.Before(time.Now()) {
		sub.EndDate = time.Now().AddDate(0, 0, int(plan.Duration))
	} else {
		sub.EndDate = sub.EndDate.AddDate(0, 0, int(plan.Duration))
	}
	sub.Status = domain.SubscriptionStatusActive

	if err := m.repo.SaveSubscription(ctx, sub); err != nil {
		m.logger.ErrorContext(ctx, "failed to renew subscription", "subscription_id", id, "error", err)
		return err
	}
	m.logger.InfoContext(ctx, "subscription renewed successfully", "subscription_id", id)
	return nil
}

// UpdatePlan updates an existing subscription plan.
func (m *SubscriptionManager) UpdatePlan(ctx context.Context, id uint64, name, desc *string, price *uint64, duration *int32, features []string, enabled *bool) (*domain.SubscriptionPlan, error) {
	plan, err := m.repo.GetPlan(ctx, id)
	if err != nil {
		return nil, err
	}
	if plan == nil {
		return nil, errors.New("plan not found")
	}

	if name != nil {
		plan.Name = *name
	}
	if desc != nil {
		plan.Description = *desc
	}
	if price != nil {
		plan.Price = *price
	}
	if duration != nil {
		plan.Duration = *duration
	}
	if features != nil {
		plan.Features = features
	}
	if enabled != nil {
		plan.Enabled = *enabled
	}

	if err := m.repo.SavePlan(ctx, plan); err != nil {
		m.logger.ErrorContext(ctx, "failed to update plan", "plan_id", id, "error", err)
		return nil, err
	}
	m.logger.InfoContext(ctx, "plan updated successfully", "plan_id", id)
	return plan, nil
}
