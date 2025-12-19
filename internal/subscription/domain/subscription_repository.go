package domain

import (
	"context"
)

// SubscriptionRepository 是订阅模块的仓储接口。
type SubscriptionRepository interface {
	// --- 订阅计划 (SubscriptionPlan methods) ---
	SavePlan(ctx context.Context, plan *SubscriptionPlan) error
	GetPlan(ctx context.Context, id uint64) (*SubscriptionPlan, error)
	ListPlans(ctx context.Context, enabledOnly bool) ([]*SubscriptionPlan, error)

	// --- 订阅 (Subscription methods) ---
	SaveSubscription(ctx context.Context, sub *Subscription) error
	GetSubscription(ctx context.Context, id uint64) (*Subscription, error)
	GetActiveSubscription(ctx context.Context, userID uint64) (*Subscription, error)
	ListSubscriptions(ctx context.Context, userID uint64, status *SubscriptionStatus, offset, limit int) ([]*Subscription, int64, error)
}
