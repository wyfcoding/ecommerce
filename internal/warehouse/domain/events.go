package domain

import (
	"time"
)

// StockDeductedEvent 库存扣减成功事件。
type StockDeductedEvent struct {
	OrderID     uint64    `json:"order_id"`
	SkuID       uint64    `json:"sku_id"`
	Quantity    int32     `json:"quantity"`
	WarehouseID uint64    `json:"warehouse_id"`
	Timestamp   time.Time `json:"timestamp"`
}

// StockRevertedEvent 库存回滚事件。
type StockRevertedEvent struct {
	OrderID     uint64    `json:"order_id"`
	SkuID       uint64    `json:"sku_id"`
	Quantity    int32     `json:"quantity"`
	WarehouseID uint64    `json:"warehouse_id"`
	Timestamp   time.Time `json:"timestamp"`
}

// WarehouseCreatedEvent 仓库创建事件。
type WarehouseCreatedEvent struct {
	WarehouseID uint64    `json:"warehouse_id"`
	Code        string    `json:"code"`
	Name        string    `json:"name"`
	Timestamp   time.Time `json:"timestamp"`
}

// StockAdjustedEvent 库存人工/系统调整事件。
type StockAdjustedEvent struct {
	WarehouseID uint64    `json:"warehouse_id"`
	SkuID       uint64    `json:"sku_id"`
	OldQty      int32     `json:"old_qty"`
	NewQty      int32     `json:"new_qty"`
	Reason      string    `json:"reason"`
	Timestamp   time.Time `json:"timestamp"`
}
