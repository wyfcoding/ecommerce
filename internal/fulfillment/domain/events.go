// Package domain 履约服务领域事件定义
package domain

import "time"

// DomainEvent 领域事件接口
type DomainEvent interface {
	EventName() string
	OccurredAt() time.Time
}

// FulfillmentCreatedEvent 履约单创建事件
type FulfillmentCreatedEvent struct {
	FulfillmentID uint64    `json:"fulfillment_id"`
	FulfillmentNo string    `json:"fulfillment_no"`
	OrderNo       string    `json:"order_no"`
	MerchantID    uint64    `json:"merchant_id"`
	Timestamp     time.Time `json:"timestamp"`
}

func (e *FulfillmentCreatedEvent) EventName() string     { return "fulfillment.created" }
func (e *FulfillmentCreatedEvent) OccurredAt() time.Time { return e.Timestamp }

// PickerAssignedEvent 拣货员分配事件
type PickerAssignedEvent struct {
	FulfillmentID uint64    `json:"fulfillment_id"`
	PickerID      uint64    `json:"picker_id"`
	PickerName    string    `json:"picker_name"`
	Timestamp     time.Time `json:"timestamp"`
}

func (e *PickerAssignedEvent) EventName() string     { return "fulfillment.picker_assigned" }
func (e *PickerAssignedEvent) OccurredAt() time.Time { return e.Timestamp }

// PickingStartedEvent 开始拣货事件
type PickingStartedEvent struct {
	FulfillmentID uint64    `json:"fulfillment_id"`
	PickerID      uint64    `json:"picker_id"`
	Timestamp     time.Time `json:"timestamp"`
}

func (e *PickingStartedEvent) EventName() string     { return "fulfillment.picking_started" }
func (e *PickingStartedEvent) OccurredAt() time.Time { return e.Timestamp }

// PickingCompletedEvent 完成拣货事件
type PickingCompletedEvent struct {
	FulfillmentID uint64    `json:"fulfillment_id"`
	PickerID      uint64    `json:"picker_id"`
	Timestamp     time.Time `json:"timestamp"`
}

func (e *PickingCompletedEvent) EventName() string     { return "fulfillment.picking_completed" }
func (e *PickingCompletedEvent) OccurredAt() time.Time { return e.Timestamp }

// PackingStartedEvent 开始打包事件
type PackingStartedEvent struct {
	FulfillmentID uint64    `json:"fulfillment_id"`
	PackerID      uint64    `json:"packer_id"`
	Timestamp     time.Time `json:"timestamp"`
}

func (e *PackingStartedEvent) EventName() string     { return "fulfillment.packing_started" }
func (e *PackingStartedEvent) OccurredAt() time.Time { return e.Timestamp }

// PackingCompletedEvent 完成打包事件
type PackingCompletedEvent struct {
	FulfillmentID uint64    `json:"fulfillment_id"`
	PackerID      uint64    `json:"packer_id"`
	PackageCount  int       `json:"package_count"`
	Timestamp     time.Time `json:"timestamp"`
}

func (e *PackingCompletedEvent) EventName() string     { return "fulfillment.packing_completed" }
func (e *PackingCompletedEvent) OccurredAt() time.Time { return e.Timestamp }

// ShipmentConfirmedEvent 确认发货事件
type ShipmentConfirmedEvent struct {
	FulfillmentID uint64    `json:"fulfillment_id"`
	TrackingNo    string    `json:"tracking_no"`
	CarrierCode   string    `json:"carrier_code"`
	Timestamp     time.Time `json:"timestamp"`
}

func (e *ShipmentConfirmedEvent) EventName() string     { return "fulfillment.shipment_confirmed" }
func (e *ShipmentConfirmedEvent) OccurredAt() time.Time { return e.Timestamp }

// FulfillmentCancelledEvent 履约取消事件
type FulfillmentCancelledEvent struct {
	FulfillmentID uint64    `json:"fulfillment_id"`
	Reason        string    `json:"reason"`
	Operator      string    `json:"operator"`
	Timestamp     time.Time `json:"timestamp"`
}

func (e *FulfillmentCancelledEvent) EventName() string     { return "fulfillment.cancelled" }
func (e *FulfillmentCancelledEvent) OccurredAt() time.Time { return e.Timestamp }
