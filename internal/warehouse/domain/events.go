package domain

import (
	"time"
)

// Event types
const (
	WarehouseCreatedEventType       = "warehouse.created"
	StockAdjustedEventType          = "warehouse.stock.adjusted"
	StockDeductedEventType          = "warehouse.stock.deducted"
	StockRevertedEventType          = "warehouse.stock.reverted"
	StockTransferCreatedEventType   = "warehouse.transfer.created"
	StockTransferCompletedEventType = "warehouse.transfer.completed"
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

// StockTransferCreatedEvent 调拨单创建事件。
type StockTransferCreatedEvent struct {
	TransferID      uint64              `json:"transfer_id"`
	TransferNo      string              `json:"transfer_no"`
	FromWarehouseID uint64              `json:"from_warehouse_id"`
	ToWarehouseID   uint64              `json:"to_warehouse_id"`
	SkuID           uint64              `json:"sku_id"`
	Quantity        int32               `json:"quantity"`
	Status          StockTransferStatus `json:"status"`
	CreatedBy       uint64              `json:"created_by"`
	Timestamp       time.Time           `json:"timestamp"`
}

// StockTransferCompletedEvent 调拨单完成事件。
type StockTransferCompletedEvent struct {
	TransferID  uint64              `json:"transfer_id"`
	TransferNo  string              `json:"transfer_no"`
	Status      StockTransferStatus `json:"status"`
	CompletedAt time.Time           `json:"completed_at"`
	Timestamp   time.Time           `json:"timestamp"`
}
