package application

import (
	"context"

	"github.com/wyfcoding/ecommerce/internal/subscription/domain"
)

// SubscriptionService 作为订阅操作的门面。
type SubscriptionService struct {
	manager *SubscriptionManager
	query   *SubscriptionQuery
}

// NewSubscriptionService creates a new SubscriptionService facade.
func NewSubscriptionService(manager *SubscriptionManager, query *SubscriptionQuery) *SubscriptionService {
	return &SubscriptionService{
		manager: manager,
		query:   query,
	}
}

// --- 写操作（委托给 Manager）---

func (s *SubscriptionService) CreatePlan(ctx context.Context, name, desc string, price uint64, duration int32, features []string) (*domain.SubscriptionPlan, error) {
	return s.manager.CreatePlan(ctx, name, desc, price, duration, features)
}

func (s *SubscriptionService) Subscribe(ctx context.Context, userID, planID uint64) (*domain.Subscription, error) {
	return s.manager.Subscribe(ctx, userID, planID)
}

func (s *SubscriptionService) Cancel(ctx context.Context, id uint64) error {
	return s.manager.Cancel(ctx, id)
}

func (s *SubscriptionService) Renew(ctx context.Context, id uint64) error {
	return s.manager.Renew(ctx, id)
}

func (s *SubscriptionService) UpdatePlan(ctx context.Context, id uint64, name, desc *string, price *uint64, duration *int32, features []string, enabled *bool) (*domain.SubscriptionPlan, error) {
	return s.manager.UpdatePlan(ctx, id, name, desc, price, duration, features, enabled)
}

// --- 读操作（委托给 Query）---

func (s *SubscriptionService) ListPlans(ctx context.Context) ([]*domain.SubscriptionPlan, error) {
	return s.query.ListPlans(ctx)
}

func (s *SubscriptionService) ListSubscriptions(ctx context.Context, userID uint64, page, pageSize int) ([]*domain.Subscription, int64, error) {
	return s.query.ListSubscriptions(ctx, userID, page, pageSize)
}

func (s *SubscriptionService) GetPlan(ctx context.Context, id uint64) (*domain.SubscriptionPlan, error) {
	return s.query.GetPlan(ctx, id)
}

func (s *SubscriptionService) GetSubscription(ctx context.Context, id uint64) (*domain.Subscription, error) {
	return s.query.GetSubscription(ctx, id)
}
