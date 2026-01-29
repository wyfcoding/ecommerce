package domain

import (
	"time"
)

// LogisticsCreatedEvent 物流单创建事件
type LogisticsCreatedEvent struct {
	LogisticsID uint      `json:"logistics_id"`
	OrderID     uint64    `json:"order_id"`
	OrderNo     string    `json:"order_no"`
	TrackingNo  string    `json:"tracking_no"`
	Timestamp   time.Time `json:"timestamp"`
}

// LogisticsStatusUpdatedEvent 物流状态更新事件
type LogisticsStatusUpdatedEvent struct {
	LogisticsID uint            `json:"logistics_id"`
	OrderID     uint64          `json:"order_id"`
	Status      LogisticsStatus `json:"status"`
	Location    string          `json:"location"`
	Description string          `json:"description"`
	Timestamp   time.Time       `json:"timestamp"`
}

// LogisticsTraceAddedEvent 物流轨迹添加事件
type LogisticsTraceAddedEvent struct {
	LogisticsID uint      `json:"logistics_id"`
	TrackingNo  string    `json:"tracking_no"`
	Location    string    `json:"location"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	Timestamp   time.Time `json:"timestamp"`
}

// RiderAssignedEvent 骑手指派事件
type RiderAssignedEvent struct {
	LogisticsID uint      `json:"logistics_id"`
	OrderID     uint64    `json:"order_id"`
	RiderID     string    `json:"rider_id"`
	Timestamp   time.Time `json:"timestamp"`
}
