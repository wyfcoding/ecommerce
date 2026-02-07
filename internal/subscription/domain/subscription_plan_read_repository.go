package domain

import "context"

// SubscriptionPlanReadRepository 定义订阅计划读模型仓储接口（Redis）。
type SubscriptionPlanReadRepository interface {
	Save(ctx context.Context, plan *SubscriptionPlan) error
	GetByID(ctx context.Context, id uint64) (*SubscriptionPlan, error)
	Delete(ctx context.Context, id uint64) error
}
