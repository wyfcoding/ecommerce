package biz

import (
	"context"
	"errors"
	"time"
)

// ErrSubscriptionPlanNotFound is a specific error for when a subscription plan is not found.
var ErrSubscriptionPlanNotFound = errors.New("subscription plan not found")

// ErrUserSubscriptionNotFound is a specific error for when a user subscription is not found.
var ErrUserSubscriptionNotFound = errors.New("user subscription not found")

// SubscriptionPlan represents a subscription plan in the business layer.
type SubscriptionPlan struct {
	ID             uint
	Name           string
	Description    string
	Price          float64
	Currency       string
	RecurrenceType string // e.g., MONTHLY, ANNUALLY
	DurationMonths int32  // 0 for indefinite
	IsActive       bool
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// UserSubscription represents a user's subscription to a plan in the business layer.
type UserSubscription struct {
	ID              uint
	UserID          string
	PlanID          string
	Status          string // e.g., ACTIVE, CANCELED, EXPIRED, PENDING_RENEWAL
	StartDate       time.Time
	EndDate         time.Time
	NextBillingDate time.Time
	PaymentMethodID string
	AutoRenew       bool
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// SubscriptionRepo defines the data storage interface for subscription data.
// The business layer depends on this interface, not on a concrete data implementation.
type SubscriptionRepo interface {
	CreateSubscriptionPlan(ctx context.Context, plan *SubscriptionPlan) (*SubscriptionPlan, error)
	GetSubscriptionPlan(ctx context.Context, id uint) (*SubscriptionPlan, error)
	ListSubscriptionPlans(ctx context.Context, activeOnly bool) ([]*SubscriptionPlan, int32, error)
	UpdateSubscriptionPlan(ctx context.Context, plan *SubscriptionPlan) (*SubscriptionPlan, error)

	CreateUserSubscription(ctx context.Context, sub *UserSubscription) (*UserSubscription, error)
	GetUserSubscription(ctx context.Context, id uint) (*UserSubscription, error)
	GetUserActiveSubscriptionByUserID(ctx context.Context, userID string) (*UserSubscription, error)
	UpdateUserSubscription(ctx context.Context, sub *UserSubscription) (*UserSubscription, error)
}

// SubscriptionUsecase is the use case for subscription-related operations.
// It orchestrates the business logic.
type SubscriptionUsecase struct {
	repo SubscriptionRepo
	// You can also inject other dependencies like a logger or a payment gateway client
}

// NewSubscriptionUsecase creates a new SubscriptionUsecase.
func NewSubscriptionUsecase(repo SubscriptionRepo) *SubscriptionUsecase {
	return &SubscriptionUsecase{repo: repo}
}

// CreateSubscriptionPlan creates a new subscription plan.
func (uc *SubscriptionUsecase) CreateSubscriptionPlan(ctx context.Context, name, description string, price float64, currency, recurrenceType string, durationMonths int32, isActive bool) (*SubscriptionPlan, error) {
	plan := &SubscriptionPlan{
		Name:           name,
		Description:    description,
		Price:          price,
		Currency:       currency,
		RecurrenceType: recurrenceType,
		DurationMonths: durationMonths,
		IsActive:       isActive,
	}
	return uc.repo.CreateSubscriptionPlan(ctx, plan)
}

// GetSubscriptionPlan retrieves a subscription plan by ID.
func (uc *SubscriptionUsecase) GetSubscriptionPlan(ctx context.Context, id uint) (*SubscriptionPlan, error) {
	return uc.repo.GetSubscriptionPlan(ctx, id)
}

// ListSubscriptionPlans lists all subscription plans.
func (uc *SubscriptionUsecase) ListSubscriptionPlans(ctx context.Context, activeOnly bool) ([]*SubscriptionPlan, int32, error) {
	return uc.repo.ListSubscriptionPlans(ctx, activeOnly)
}

// CreateUserSubscription creates a new user subscription.
func (uc *SubscriptionUsecase) CreateUserSubscription(ctx context.Context, userID, planID, paymentMethodID string, autoRenew bool) (*UserSubscription, error) {
	plan, err := uc.repo.GetSubscriptionPlan(ctx, 0) // Assuming planID is passed as string, need to convert to uint or get by string ID
	if err != nil {
		return nil, err
	}

	// Calculate start and end dates
	now := time.Now()
	endDate := time.Time{}
	if plan.DurationMonths > 0 {
		endDate = now.AddDate(0, int(plan.DurationMonths), 0)
	}

	subscription := &UserSubscription{
		UserID:          userID,
		PlanID:          planID,
		Status:          "ACTIVE",
		StartDate:       now,
		EndDate:         endDate,
		NextBillingDate: now.AddDate(0, 1, 0), // Assuming monthly billing for now
		PaymentMethodID: paymentMethodID,
		AutoRenew:       autoRenew,
	}
	return uc.repo.CreateUserSubscription(ctx, subscription)
}

// GetUserSubscription retrieves a user's subscription by ID.
func (uc *SubscriptionUsecase) GetUserSubscription(ctx context.Context, id uint) (*UserSubscription, error) {
	return uc.repo.GetUserSubscription(ctx, id)
}

// CancelUserSubscription cancels a user's subscription.
func (uc *SubscriptionUsecase) CancelUserSubscription(ctx context.Context, id uint) (*UserSubscription, error) {
	sub, err := uc.repo.GetUserSubscription(ctx, id)
	if err != nil {
		return nil, err
	}
	sub.Status = "CANCELED"
	sub.AutoRenew = false
	return uc.repo.UpdateUserSubscription(ctx, sub)
}

// UpdateUserSubscription updates a user's subscription.
func (uc *SubscriptionUsecase) UpdateUserSubscription(ctx context.Context, id uint, planID, paymentMethodID string, autoRenew bool) (*UserSubscription, error) {
	sub, err := uc.repo.GetUserSubscription(ctx, id)
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

	return uc.repo.UpdateUserSubscription(ctx, sub)
}
