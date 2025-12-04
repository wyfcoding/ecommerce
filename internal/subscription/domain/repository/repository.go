package repository

import (
	"context"
	"github.com/wyfcoding/ecommerce/internal/subscription/domain/entity" // 导入订阅领域的实体定义。
)

// SubscriptionRepository 是订阅模块的仓储接口。
// 它定义了对订阅计划和订阅实体进行数据持久化操作的契约。
// 仓储接口属于领域层，旨在将领域逻辑与数据存储的实现细节解耦。
type SubscriptionRepository interface {
	// --- 订阅计划 (SubscriptionPlan methods) ---

	// SavePlan 将订阅计划实体保存到数据存储中。
	// 如果计划已存在，则更新；如果不存在，则创建。
	// ctx: 上下文。
	// plan: 待保存的订阅计划实体。
	SavePlan(ctx context.Context, plan *entity.SubscriptionPlan) error
	// GetPlan 根据ID获取订阅计划实体。
	GetPlan(ctx context.Context, id uint64) (*entity.SubscriptionPlan, error)
	// ListPlans 列出所有订阅计划实体。
	// enabledOnly: 布尔值，如果为true，则只列出启用的计划。
	ListPlans(ctx context.Context, enabledOnly bool) ([]*entity.SubscriptionPlan, error)

	// --- 订阅 (Subscription methods) ---

	// SaveSubscription 将订阅实体保存到数据存储中。
	SaveSubscription(ctx context.Context, sub *entity.Subscription) error
	// GetSubscription 根据ID获取订阅实体。
	GetSubscription(ctx context.Context, id uint64) (*entity.Subscription, error)
	// GetActiveSubscription 获取指定用户ID的活跃订阅实体。
	GetActiveSubscription(ctx context.Context, userID uint64) (*entity.Subscription, error)
	// ListSubscriptions 列出指定用户ID的所有订阅实体，支持通过状态过滤和分页。
	ListSubscriptions(ctx context.Context, userID uint64, status *entity.SubscriptionStatus, offset, limit int) ([]*entity.Subscription, int64, error)
}
