package domain

import (
	"context"
	"time"
)

// SubscriptionRepository 是订阅模块的写模型仓储接口。
type SubscriptionRepository interface {
	// --- tx helpers ---
	BeginTx(ctx context.Context) any
	CommitTx(tx any) error
	RollbackTx(tx any) error
	WithTx(ctx context.Context, fn func(tx any) error) error

	// --- 订阅计划 (SubscriptionPlan methods) ---
	SavePlan(ctx context.Context, plan *SubscriptionPlan) error
	SavePlanInTx(ctx context.Context, tx any, plan *SubscriptionPlan) error
	GetPlan(ctx context.Context, id uint64) (*SubscriptionPlan, error)
	ListPlans(ctx context.Context, enabledOnly bool) ([]*SubscriptionPlan, error)
	DeletePlan(ctx context.Context, id uint64) error
	DeletePlanInTx(ctx context.Context, tx any, id uint64) error

	// --- 订阅 (Subscription methods) ---
	SaveSubscription(ctx context.Context, sub *Subscription) error
	SaveSubscriptionInTx(ctx context.Context, tx any, sub *Subscription) error
	GetSubscription(ctx context.Context, id uint64) (*Subscription, error)
	GetActiveSubscription(ctx context.Context, userID uint64) (*Subscription, error)
	ListSubscriptions(ctx context.Context, query *SubscriptionQuery) ([]*Subscription, int64, error)
	DeleteSubscription(ctx context.Context, id uint64) error
	DeleteSubscriptionInTx(ctx context.Context, tx any, id uint64) error
}

// SubscriptionQuery 定义订阅记录的查询条件。
type SubscriptionQuery struct {
	UserID    uint64
	PlanID    uint64
	Status    *SubscriptionStatus
	AutoRenew *bool
	StartTime time.Time
	EndTime   time.Time
	Page      int
	PageSize  int
}
