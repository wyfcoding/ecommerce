package domain

import "time"

// CustomerRegisteredEvent 客户注册事件。
type CustomerRegisteredEvent struct {
	CustomerID uint64    `json:"customer_id"`
	Email      string    `json:"email"`
	Timestamp  time.Time `json:"timestamp"`
}

// CustomerProfileUpdatedEvent 客户资料更新事件。
type CustomerProfileUpdatedEvent struct {
	CustomerID uint64    `json:"customer_id"`
	Timestamp  time.Time `json:"timestamp"`
}

// CustomerAccountStatusChangedEvent 客户账号状态变更事件。
type CustomerAccountStatusChangedEvent struct {
	CustomerID uint64    `json:"customer_id"`
	OldStatus  string    `json:"old_status"`
	NewStatus  string    `json:"new_status"`
	Timestamp  time.Time `json:"timestamp"`
}
