package domain

import (
	"time"

	"github.com/wyfcoding/pkg/eventsourcing"
)

// Event types
const (
	StockLockedEventType   = "inventory.stock.locked"
	StockUnlockedEventType = "inventory.stock.unlocked"
	StockDeductedEventType = "inventory.stock.deducted"
	StockAddedEventType    = "inventory.stock.added"
	StockWarningEventType  = "inventory.stock.warning"
)

// StockLockedEvent 库存锁定事件
type StockLockedEvent struct {
	eventsourcing.BaseEvent
	SkuID      uint64    `json:"sku_id"`
	Quantity   int32     `json:"quantity"`
	Reason     string    `json:"reason"`
	Timestamp  time.Time `json:"timestamp"`
}

func (e *StockLockedEvent) EventType() string { return StockLockedEventType }

// StockUnlockedEvent 库存解锁事件
type StockUnlockedEvent struct {
	eventsourcing.BaseEvent
	SkuID     uint64    `json:"sku_id"`
	Quantity  int32     `json:"quantity"`
	Reason    string    `json:"reason"`
	Timestamp time.Time `json:"timestamp"`
}

func (e *StockUnlockedEvent) EventType() string { return StockUnlockedEventType }

// StockDeductedEvent 库存扣减事件
type StockDeductedEvent struct {
	eventsourcing.BaseEvent
	SkuID     uint64    `json:"sku_id"`
	Quantity  int32     `json:"quantity"`
	Reason    string    `json:"reason"`
	Timestamp time.Time `json:"timestamp"`
}

func (e *StockDeductedEvent) EventType() string { return StockDeductedEventType }

// StockAddedEvent 库存增加事件
type StockAddedEvent struct {
	eventsourcing.BaseEvent
	SkuID     uint64    `json:"sku_id"`
	Quantity  int32     `json:"quantity"`
	Reason    string    `json:"reason"`
	Timestamp time.Time `json:"timestamp"`
}

func (e *StockAddedEvent) EventType() string { return StockAddedEventType }

// StockWarningEvent 库存预警事件
type StockWarningEvent struct {
	eventsourcing.BaseEvent
	SkuID          uint64    `json:"sku_id"`
	AvailableStock int32     `json:"available_stock"`
	Threshold      int32     `json:"threshold"`
	Timestamp      time.Time `json:"timestamp"`
}

func (e *StockWarningEvent) EventType() string { return StockWarningEventType }
