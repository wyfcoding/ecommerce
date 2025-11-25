package repository

import (
	"context"
	"github.com/wyfcoding/ecommerce/internal/subscription/domain/entity"
)

// SubscriptionRepository 订阅仓储接口
type SubscriptionRepository interface {
	// 订阅计划
	SavePlan(ctx context.Context, plan *entity.SubscriptionPlan) error
	GetPlan(ctx context.Context, id uint64) (*entity.SubscriptionPlan, error)
	ListPlans(ctx context.Context, enabledOnly bool) ([]*entity.SubscriptionPlan, error)

	// 订阅
	SaveSubscription(ctx context.Context, sub *entity.Subscription) error
	GetSubscription(ctx context.Context, id uint64) (*entity.Subscription, error)
	GetActiveSubscription(ctx context.Context, userID uint64) (*entity.Subscription, error)
	ListSubscriptions(ctx context.Context, userID uint64, status *entity.SubscriptionStatus, offset, limit int) ([]*entity.Subscription, int64, error)
}
