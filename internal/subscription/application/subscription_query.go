package application

import (
	"context"

	"github.com/wyfcoding/ecommerce/internal/subscription/domain"
)

// SubscriptionQuery 处理订阅的读操作。
type SubscriptionQuery struct {
	repo domain.SubscriptionRepository
}

// NewSubscriptionQuery creates a new SubscriptionQuery instance.
func NewSubscriptionQuery(repo domain.SubscriptionRepository) *SubscriptionQuery {
	return &SubscriptionQuery{
		repo: repo,
	}
}

// ListPlans retrieves all enabled subscription plans.
func (q *SubscriptionQuery) ListPlans(ctx context.Context) ([]*domain.SubscriptionPlan, error) {
	return q.repo.ListPlans(ctx, true)
}

// ListSubscriptions 获取用户的订阅列表。
func (q *SubscriptionQuery) ListSubscriptions(ctx context.Context, userID uint64, page, pageSize int) ([]*domain.Subscription, int64, error) {
	offset := (page - 1) * pageSize
	return q.repo.ListSubscriptions(ctx, userID, nil, offset, pageSize)
}

// GetPlan retrieves a plan by ID.
func (q *SubscriptionQuery) GetPlan(ctx context.Context, id uint64) (*domain.SubscriptionPlan, error) {
	return q.repo.GetPlan(ctx, id)
}

// GetSubscription retrieves a subscription by ID.
func (q *SubscriptionQuery) GetSubscription(ctx context.Context, id uint64) (*domain.Subscription, error) {
	return q.repo.GetSubscription(ctx, id)
}
