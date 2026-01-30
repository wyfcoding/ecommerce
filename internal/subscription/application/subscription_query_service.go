package application

import (
	"context"

	"github.com/wyfcoding/ecommerce/internal/subscription/domain"
)

// SubscriptionQueryService 处理订阅的读操作。
type SubscriptionQueryService struct {
	repo domain.SubscriptionRepository
}

// NewSubscriptionQueryService creates a new SubscriptionQueryService instance.
func NewSubscriptionQueryService(repo domain.SubscriptionRepository) *SubscriptionQueryService {
	return &SubscriptionQueryService{
		repo: repo,
	}
}

// ListPlans retrieves all enabled subscription plans.
func (q *SubscriptionQueryService) ListPlans(ctx context.Context) ([]*domain.SubscriptionPlan, error) {
	return q.repo.ListPlans(ctx, true)
}

// ListSubscriptions 获取用户的订阅列表。
func (q *SubscriptionQueryService) ListSubscriptions(ctx context.Context, userID uint64, page, pageSize int) ([]*domain.Subscription, int64, error) {
	offset := (page - 1) * pageSize
	return q.repo.ListSubscriptions(ctx, userID, nil, offset, pageSize)
}

// GetPlan retrieves a plan by ID.
func (q *SubscriptionQueryService) GetPlan(ctx context.Context, id uint64) (*domain.SubscriptionPlan, error) {
	return q.repo.GetPlan(ctx, id)
}

// GetSubscription retrieves a subscription by ID.
func (q *SubscriptionQueryService) GetSubscription(ctx context.Context, id uint64) (*domain.Subscription, error) {
	return q.repo.GetSubscription(ctx, id)
}
