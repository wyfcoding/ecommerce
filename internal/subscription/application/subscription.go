package application

import (
	"context"

	"github.com/wyfcoding/ecommerce/internal/subscription/domain"
)

// SubscriptionService 作为订阅操作的门面。
type SubscriptionService struct {
	Command *SubscriptionCommandService
	Query   *SubscriptionQueryService
}

// NewSubscriptionService 创建订阅服务门面实例。
func NewSubscriptionService(command *SubscriptionCommandService, query *SubscriptionQueryService) *SubscriptionService {
	return &SubscriptionService{
		Command: command,
		Query:   query,
	}
}

// --- 写操作（委托给 Manager）---

// CreatePlan 创建一个新的订阅计划（套餐）。
func (s *SubscriptionService) CreatePlan(ctx context.Context, name, desc string, price uint64, duration int32, features []string) (*domain.SubscriptionPlan, error) {
	return s.Command.CreatePlan(ctx, name, desc, price, duration, features)
}

// Subscribe 为用户开启一个新的订阅。
func (s *SubscriptionService) Subscribe(ctx context.Context, userID, planID uint64) (*domain.Subscription, error) {
	return s.Command.Subscribe(ctx, userID, planID)
}

// Cancel 取消用户的有效订阅（通常是关闭自动续费）。
func (s *SubscriptionService) Cancel(ctx context.Context, id uint64) error {
	return s.Command.Cancel(ctx, id)
}

// Renew 为即将到期或已到期的订阅进行续费。
func (s *SubscriptionService) Renew(ctx context.Context, id uint64) error {
	return s.Command.Renew(ctx, id)
}

// UpdatePlan 更新现有订阅计划的详细信息。
func (s *SubscriptionService) UpdatePlan(ctx context.Context, id uint64, name, desc *string, price *uint64, duration *int32, features []string, enabled *bool) (*domain.SubscriptionPlan, error) {
	return s.Command.UpdatePlan(ctx, id, name, desc, price, duration, features, enabled)
}

// --- 读操作（委托给 Query）---

// ListPlans 列出所有可用的订阅计划。
func (s *SubscriptionService) ListPlans(ctx context.Context) ([]*domain.SubscriptionPlan, error) {
	return s.Query.ListPlans(ctx)
}

// ListSubscriptions 获取指定用户的订阅历史记录（分页）。
func (s *SubscriptionService) ListSubscriptions(ctx context.Context, userID uint64, page, pageSize int) ([]*domain.Subscription, int64, error) {
	return s.Query.ListSubscriptions(ctx, userID, page, pageSize)
}

// GetPlan 获取指定ID的订阅计划详情。
func (s *SubscriptionService) GetPlan(ctx context.Context, id uint64) (*domain.SubscriptionPlan, error) {
	return s.Query.GetPlan(ctx, id)
}

// GetSubscription 获取指定ID的订阅详情。
func (s *SubscriptionService) GetSubscription(ctx context.Context, id uint64) (*domain.Subscription, error) {
	return s.Query.GetSubscription(ctx, id)
}
