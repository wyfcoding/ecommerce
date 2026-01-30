package domain

import "time"

// ChannelOrderCreatedEvent 渠道订单创建事件。
type ChannelOrderCreatedEvent struct {
	OrderID     uint64    `json:"order_id"`
	ChannelID   uint64    `json:"channel_id"`
	ExternalID  string    `json:"external_id"`
	TotalAmount int64     `json:"total_amount"`
	Timestamp   time.Time `json:"timestamp"`
}

// ChannelStatusUpdatedEvent 渠道状态更新事件。
type ChannelStatusUpdatedEvent struct {
	ChannelID uint64    `json:"channel_id"`
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
}
