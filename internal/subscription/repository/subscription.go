package repository

import (
	"context"

	"ecommerce/internal/subscription/model"
)

// SubscriptionRepo defines the data storage interface for subscription data.
// The business layer depends on this interface, not on a concrete data implementation.
type SubscriptionRepo interface {
	CreateSubscriptionPlan(ctx context.Context, plan *model.SubscriptionPlan) (*model.SubscriptionPlan, error)
	GetSubscriptionPlan(ctx context.Context, id uint) (*model.SubscriptionPlan, error)
	ListSubscriptionPlans(ctx context.Context, activeOnly bool) ([]*model.SubscriptionPlan, int32, error)
	UpdateSubscriptionPlan(ctx context.Context, plan *model.SubscriptionPlan) (*model.SubscriptionPlan, error)

	CreateUserSubscription(ctx context.Context, sub *model.UserSubscription) (*model.UserSubscription, error)
	GetUserSubscription(ctx context.Context, id uint) (*model.UserSubscription, error)
	GetUserActiveSubscriptionByUserID(ctx context.Context, userID string) (*model.UserSubscription, error)
	UpdateUserSubscription(ctx context.Context, sub *model.UserSubscription) (*model.UserSubscription, error)
}