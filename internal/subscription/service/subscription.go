package service

import (
	"context"
	"errors"
	"time"

	"ecommerce/internal/subscription/model"
	"ecommerce/internal/subscription/repository"
)

// ErrSubscriptionPlanNotFound is a specific error for when a subscription plan is not found.
var ErrSubscriptionPlanNotFound = errors.New("subscription plan not found")

// ErrUserSubscriptionNotFound is a specific error for when a user subscription is not found.
var ErrUserSubscriptionNotFound = errors.New("user subscription not found")

// SubscriptionService is the use case for subscription-related operations.
// It orchestrates the business logic.
type SubscriptionService struct {
	repo repository.SubscriptionRepo
	// You can also inject other dependencies like a logger or a payment gateway client
}

// NewSubscriptionService creates a new SubscriptionService.
func NewSubscriptionService(repo repository.SubscriptionRepo) *SubscriptionService {
	return &SubscriptionService{repo: repo}
}

// CreateSubscriptionPlan creates a new subscription plan.
func (s *SubscriptionService) CreateSubscriptionPlan(ctx context.Context, name, description string, price float64, currency, recurrenceType string, durationMonths int32, isActive bool) (*model.SubscriptionPlan, error) {
	plan := &model.SubscriptionPlan{
		Name:           name,
		Description:    description,
		Price:          price,
		Currency:       currency,
		RecurrenceType: recurrenceType,
		DurationMonths: durationMonths,
		IsActive:       isActive,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	return s.repo.CreateSubscriptionPlan(ctx, plan)
}

// GetSubscriptionPlan retrieves a subscription plan by ID.
func (s *SubscriptionService) GetSubscriptionPlan(ctx context.Context, id uint) (*model.SubscriptionPlan, error) {
	return s.repo.GetSubscriptionPlan(ctx, id)
}

// ListSubscriptionPlans lists all subscription plans.
func (s *SubscriptionService) ListSubscriptionPlans(ctx context.Context, activeOnly bool) ([]*model.SubscriptionPlan, int32, error) {
	return s.repo.ListSubscriptionPlans(ctx, activeOnly)
}

// CreateUserSubscription creates a new user subscription.
func (s *SubscriptionService) CreateUserSubscription(ctx context.Context, userID, planID, paymentMethodID string, autoRenew bool) (*model.UserSubscription, error) {
	plan, err := s.repo.GetSubscriptionPlan(ctx, 0) // Assuming planID is passed as string, need to convert to uint or get by string ID
	if err != nil {
		return nil, err
	}

	// Calculate start and end dates
	now := time.Now()
	endDate := time.Time{}
	if plan.DurationMonths > 0 {
		endDate = now.AddDate(0, int(plan.DurationMonths), 0)
	}

	subscription := &model.UserSubscription{
		UserID:          userID,
		PlanID:          planID,
		Status:          "ACTIVE",
		StartDate:       now,
		EndDate:         endDate,
		NextBillingDate: now.AddDate(0, 1, 0), // Assuming monthly billing for now
		PaymentMethodID: paymentMethodID,
		AutoRenew:       autoRenew,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
	return s.repo.CreateUserSubscription(ctx, subscription)
}

// GetUserSubscription retrieves a user's subscription by ID.
func (s *SubscriptionService) GetUserSubscription(ctx context.Context, id uint) (*model.UserSubscription, error) {
	return s.repo.GetUserSubscription(ctx, id)
}

// CancelUserSubscription cancels a user's subscription.
func (s *SubscriptionService) CancelUserSubscription(ctx context.Context, id uint) (*model.UserSubscription, error) {
	sub, err := s.repo.GetUserSubscription(ctx, id)
	if err != nil {
		return nil, err
	}
	sub.Status = "CANCELED"
	sub.AutoRenew = false
	return s.repo.UpdateUserSubscription(ctx, sub)
}

// UpdateUserSubscription updates a user's subscription.
func (s *SubscriptionService) UpdateUserSubscription(ctx context.Context, id uint, planID, paymentMethodID string, autoRenew bool) (*model.UserSubscription, error) {
	sub, err := s.repo.GetUserSubscription(ctx, id)
	if err != nil {
		return nil, err
	}

	if planID != "" && sub.PlanID != planID {
		// Logic to handle plan change
		sub.PlanID = planID
		// Recalculate dates, prorate, etc.
	}

	if paymentMethodID != "" {
		sub.PaymentMethodID = paymentMethodID
	}
	sub.AutoRenew = autoRenew

	return s.repo.UpdateUserSubscription(ctx, sub)
}
