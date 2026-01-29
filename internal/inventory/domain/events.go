package domain

import (
	"time"
)

// StockLockedEvent 库存锁定事件
type StockLockedEvent struct {
	SkuID     uint64    `json:"sku_id"`
	Quantity  int32     `json:"quantity"`
	Reason    string    `json:"reason"`
	Timestamp time.Time `json:"timestamp"`
}

// StockUnlockedEvent 库存解锁事件
type StockUnlockedEvent struct {
	SkuID     uint64    `json:"sku_id"`
	Quantity  int32     `json:"quantity"`
	Reason    string    `json:"reason"`
	Timestamp time.Time `json:"timestamp"`
}

// StockDeductedEvent 库存扣减事件
type StockDeductedEvent struct {
	SkuID     uint64    `json:"sku_id"`
	Quantity  int32     `json:"quantity"`
	Reason    string    `json:"reason"`
	Timestamp time.Time `json:"timestamp"`
}

// StockAddedEvent 库存增加事件
type StockAddedEvent struct {
	SkuID     uint64    `json:"sku_id"`
	Quantity  int32     `json:"quantity"`
	Reason    string    `json:"reason"`
	Timestamp time.Time `json:"timestamp"`
}

// StockWarningEvent 库存预警事件
type StockWarningEvent struct {
	SkuID          uint64    `json:"sku_id"`
	AvailableStock int32     `json:"available_stock"`
	Threshold      int32     `json:"threshold"`
	Timestamp      time.Time `json:"timestamp"`
}
