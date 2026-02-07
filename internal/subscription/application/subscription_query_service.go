package application

import (
	"context"
	"log/slog"

	"github.com/wyfcoding/ecommerce/internal/subscription/domain"
)

// SubscriptionQueryService 处理订阅的读操作。
type SubscriptionQueryService struct {
	repo                   domain.SubscriptionRepository
	planReadRepo           domain.SubscriptionPlanReadRepository
	subscriptionReadRepo   domain.SubscriptionReadRepository
	subscriptionSearchRepo domain.SubscriptionSearchRepository
	logger                 *slog.Logger
}

// NewSubscriptionQueryService creates a new SubscriptionQueryService instance.
func NewSubscriptionQueryService(
	repo domain.SubscriptionRepository,
	planReadRepo domain.SubscriptionPlanReadRepository,
	subscriptionReadRepo domain.SubscriptionReadRepository,
	subscriptionSearchRepo domain.SubscriptionSearchRepository,
	logger *slog.Logger,
) *SubscriptionQueryService {
	return &SubscriptionQueryService{
		repo:                   repo,
		planReadRepo:           planReadRepo,
		subscriptionReadRepo:   subscriptionReadRepo,
		subscriptionSearchRepo: subscriptionSearchRepo,
		logger:                 logger,
	}
}

// ListPlans retrieves all enabled subscription plans.
func (q *SubscriptionQueryService) ListPlans(ctx context.Context) ([]*domain.SubscriptionPlan, error) {
	return q.repo.ListPlans(ctx, true)
}

// ListSubscriptions 获取用户的订阅列表。
func (q *SubscriptionQueryService) ListSubscriptions(ctx context.Context, userID uint64, page, pageSize int) ([]*domain.Subscription, int64, error) {
	query := &domain.SubscriptionQuery{
		UserID:   userID,
		Page:     page,
		PageSize: pageSize,
	}
	return q.SearchSubscriptions(ctx, query)
}

// GetPlan retrieves a plan by ID.
func (q *SubscriptionQueryService) GetPlan(ctx context.Context, id uint64) (*domain.SubscriptionPlan, error) {
	if q.planReadRepo != nil {
		if cached, err := q.planReadRepo.GetByID(ctx, id); err == nil && cached != nil {
			return cached, nil
		}
	}
	plan, err := q.repo.GetPlan(ctx, id)
	if err != nil {
		return nil, err
	}
	if plan != nil && q.planReadRepo != nil {
		_ = q.planReadRepo.Save(ctx, plan)
	}
	return plan, nil
}

// GetSubscription retrieves a subscription by ID.
func (q *SubscriptionQueryService) GetSubscription(ctx context.Context, id uint64) (*domain.Subscription, error) {
	if q.subscriptionReadRepo != nil {
		if cached, err := q.subscriptionReadRepo.GetByID(ctx, id); err == nil && cached != nil {
			return cached, nil
		}
	}
	sub, err := q.repo.GetSubscription(ctx, id)
	if err != nil {
		return nil, err
	}
	if sub != nil && q.subscriptionReadRepo != nil {
		_ = q.subscriptionReadRepo.Save(ctx, sub)
	}
	return sub, nil
}

// SearchSubscriptions 查询订阅记录（优先 ES）。
func (q *SubscriptionQueryService) SearchSubscriptions(ctx context.Context, query *domain.SubscriptionQuery) ([]*domain.Subscription, int64, error) {
	page := 1
	pageSize := 20
	if query != nil {
		if query.Page > 0 {
			page = query.Page
		}
		if query.PageSize > 0 {
			pageSize = query.PageSize
		}
	}
	offset := (page - 1) * pageSize

	if q.subscriptionSearchRepo != nil {
		list, total, err := q.subscriptionSearchRepo.Search(ctx, query, offset, pageSize)
		if err == nil {
			return list, total, nil
		}
		q.logger.WarnContext(ctx, "subscription search fallback to mysql", "error", err)
	}
	return q.repo.ListSubscriptions(ctx, query)
}
