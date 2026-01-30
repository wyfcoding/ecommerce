package domain

import "time"

// GroupBuyCreatedEvent 团购创建事件。
type GroupBuyCreatedEvent struct {
	GroupBuyID uint64    `json:"group_buy_id"`
	ProductID  uint64    `json:"product_id"`
	CreatorID  uint64    `json:"creator_id"`
	Timestamp  time.Time `json:"timestamp"`
}

// GroupBuyJoinedEvent 用户加入团购事件。
type GroupBuyJoinedEvent struct {
	GroupBuyID uint64    `json:"group_buy_id"`
	UserID     uint64    `json:"user_id"`
	OrderID    uint64    `json:"order_id"`
	Timestamp  time.Time `json:"timestamp"`
}

// GroupBuyCompletedEvent 团购成团事件。
type GroupBuyCompletedEvent struct {
	GroupBuyID uint64    `json:"group_buy_id"`
	Timestamp  time.Time `json:"timestamp"`
}

// GroupBuyCancelledEvent 团购取消/过期事件。
type GroupBuyCancelledEvent struct {
	GroupBuyID uint64    `json:"group_buy_id"`
	Reason     string    `json:"reason"`
	Timestamp  time.Time `json:"timestamp"`
}
