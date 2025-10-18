package model

import (
	"time"

	"gorm.io/gorm"
)

// ShipmentStatus 定义了货运单的状态
type ShipmentStatus string

const (
	StatusPending     ShipmentStatus = "PENDING"     // 待处理 (等待创建货运)
	StatusCreated     ShipmentStatus = "CREATED"     // 已创建 (已获取运单号，待揽收)
	StatusInTransit   ShipmentStatus = "IN_TRANSIT"   // 运输中
	StatusDelivered   ShipmentStatus = "DELIVERED"   // 已妥投
	StatusFailure     ShipmentStatus = "FAILURE"     // 投递失败
	StatusReturned    ShipmentStatus = "RETURNED"    // 已退回
)

// Shipment 货运单模型
// 记录了与一次发货相关的所有信息
type Shipment struct {
	ID             uint           `gorm:"primarykey" json:"id"`
	OrderSN        string         `gorm:"type:varchar(100);uniqueIndex;not null" json:"order_sn"` // 关联的订单号
	Carrier        string         `gorm:"type:varchar(50);not null" json:"carrier"`                // 承运商 (e.g., "FedEx", "UPS")
	TrackingNumber string         `gorm:"type:varchar(100);index;not null" json:"tracking_number"`  // 运单号
	Status         ShipmentStatus `gorm:"type:varchar(20);not null" json:"status"`                 // 当前状态
	LabelURL       string         `gorm:"type:varchar(255)" json:"label_url"`                      // 运单标签的 URL
	EstimatedCost  float64        `gorm:"type:decimal(10,2)" json:"estimated_cost"`              // 预估运费
	ActualCost     float64        `gorm:"type:decimal(10,2)" json:"actual_cost"`                 // 实际运费
	ShippedAt      *time.Time     `json:"shipped_at"`                                              // 揽收时间
	DeliveredAt    *time.Time     `json:"delivered_at"`                                            // 妥投时间
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`

	TrackingEvents []TrackingEvent `gorm:"foreignKey:ShipmentID" json:"tracking_events"` // 追踪事件历史
}

// TrackingEvent 追踪事件模型
// 记录了从承运商获取的每一个追踪状态更新
type TrackingEvent struct {
	ID          uint      `gorm:"primarykey" json:"id"`
	ShipmentID  uint      `gorm:"not null;index" json:"shipment_id"` // 关联的货运单ID
	Description string    `gorm:"type:varchar(255);not null" json:"description"` // 事件描述 (e.g., "Arrived at Sort Facility")
	Location    string    `gorm:"type:varchar(255)" json:"location"`     // 事件发生地点
	OccurredAt  time.Time `gorm:"not null" json:"occurred_at"`         // 事件发生时间
	CreatedAt   time.Time `json:"created_at"`
}

// TableName 自定义表名
func (Shipment) TableName() string {
	return "shipments"
}

func (TrackingEvent) TableName() string {
	return "tracking_events"
}
