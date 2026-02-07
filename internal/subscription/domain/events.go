package domain

import "time"

const (
	SubscriptionPlanCreatedEventType = "subscription.plan.created"
	SubscriptionPlanUpdatedEventType = "subscription.plan.updated"
	SubscriptionPlanDeletedEventType = "subscription.plan.deleted"

	SubscriptionCreatedEventType   = "subscription.created"
	SubscriptionCancelledEventType = "subscription.cancelled"
	SubscriptionRenewedEventType   = "subscription.renewed"
)

// SubscriptionCreatedEvent 订阅创建事件。
type SubscriptionCreatedEvent struct {
	SubscriptionID uint64    `json:"subscription_id"`
	UserID         uint64    `json:"user_id"`
	PlanID         uint64    `json:"plan_id"`
	Timestamp      time.Time `json:"timestamp"`
}

// SubscriptionCancelledEvent 订阅取消事件。
type SubscriptionCancelledEvent struct {
	SubscriptionID uint64    `json:"subscription_id"`
	UserID         uint64    `json:"user_id"`
	Timestamp      time.Time `json:"timestamp"`
}

// SubscriptionRenewedEvent 订阅续订事件。
type SubscriptionRenewedEvent struct {
	SubscriptionID uint64    `json:"subscription_id"`
	UserID         uint64    `json:"user_id"`
	NewExpiryDate  time.Time `json:"new_expiry_date"`
	Timestamp      time.Time `json:"timestamp"`
}

// SubscriptionPlanCreatedEvent 订阅计划创建事件。
type SubscriptionPlanCreatedEvent struct {
	PlanID    uint64    `json:"plan_id"`
	Name      string    `json:"name"`
	Timestamp time.Time `json:"timestamp"`
}

// SubscriptionPlanUpdatedEvent 订阅计划更新事件。
type SubscriptionPlanUpdatedEvent struct {
	PlanID    uint64    `json:"plan_id"`
	Name      string    `json:"name"`
	Timestamp time.Time `json:"timestamp"`
}

// SubscriptionPlanDeletedEvent 订阅计划删除事件。
type SubscriptionPlanDeletedEvent struct {
	PlanID    uint64    `json:"plan_id"`
	Timestamp time.Time `json:"timestamp"`
}
