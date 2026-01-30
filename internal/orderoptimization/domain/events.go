package domain

import "time"

// OrderOptimizedEvent 订单优化完成事件。
type OrderOptimizedEvent struct {
	OrderID   uint64    `json:"order_id"`
	UserID    uint64    `json:"user_id"`
	Strategy  string    `json:"strategy"`
	Benefit   string    `json:"benefit"`
	Timestamp time.Time `json:"timestamp"`
}
