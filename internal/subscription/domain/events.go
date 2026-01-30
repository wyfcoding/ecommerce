package domain

import "time"

// SubscriptionCreatedEvent 订阅创建事件。
type SubscriptionCreatedEvent struct {
	SubscriptionID uint64    `json:"subscription_id"`
	UserID         uint64    `json:"user_id"`
	Type           string    `json:"type"`
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
