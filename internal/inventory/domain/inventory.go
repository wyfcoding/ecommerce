package domain

import (
	"errors"
	"fmt"
	"time"

	"github.com/wyfcoding/pkg/eventsourcing"
)

// 定义Inventory模块的业务错误。
var (
	ErrInsufficientStock = errors.New("insufficient stock")        // 库存不足以完成操作。
	ErrNegativeQuantity  = errors.New("quantity must be positive") // 操作数量必须为正数。
)

// InventoryStatus 定义了库存的生命周期状态。
type InventoryStatus int

const (
	InventoryStatusNormal     InventoryStatus = 1 // 正常：库存充足，无预警。
	InventoryStatusLocked     InventoryStatus = 2 // 已锁定：部分库存被锁定，等待订单确认。
	InventoryStatusWarning    InventoryStatus = 3 // 预警：库存量已低于警告阈值。
	InventoryStatusOutOfStock InventoryStatus = 4 // 缺货：库存已用完。
)

// Inventory 实体是库存模块的聚合根。
type Inventory struct {
	ID        uint      `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	eventsourcing.AggregateRoot
	SkuID            uint64          `json:"sku_id"`
	ProductID        uint64          `json:"product_id"`
	WarehouseID      uint64          `json:"warehouse_id"`
	AvailableStock   int32           `json:"available_stock"`
	LockedStock      int32           `json:"locked_stock"`
	TotalStock       int32           `json:"total_stock"`
	Status           InventoryStatus `json:"status"`
	WarningThreshold int32           `json:"warning_threshold"`
	PersistenceVer   int64           `json:"version"` // 改名为 PersistenceVer 避免与 AggregateRoot.Version() 冲突
}

// GetID 返回聚合标识。
func (inv *Inventory) GetID() string {
	return inv.AggregateRoot.ID()
}

// Apply 实现 eventsourcing.EventApplier 接口。
func (inv *Inventory) Apply(event eventsourcing.DomainEvent) {
	switch e := event.(type) {
	case *StockLockedEvent:
		inv.AvailableStock -= e.Quantity
		inv.LockedStock += e.Quantity
	case *StockUnlockedEvent:
		inv.AvailableStock += e.Quantity
		inv.LockedStock -= e.Quantity
	case *StockDeductedEvent:
		inv.AvailableStock -= e.Quantity
		inv.TotalStock -= e.Quantity
	case *StockAddedEvent:
		inv.AvailableStock += e.Quantity
		inv.TotalStock += e.Quantity
	}
	inv.updateStatus()
	inv.SetVersion(event.Version())
}

// InventoryLog 实体代表库存的一次操作日志。
type InventoryLog struct {
	ID             uint      `json:"id"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	InventoryID    uint64    `json:"inventory_id"`
	SkuID          uint64    `json:"sku_id"`
	Action         string    `json:"action"`
	ChangeQuantity int32     `json:"change_quantity"`
	OldAvailable   int32     `json:"old_available"`
	NewAvailable   int32     `json:"new_available"`
	OldLocked      int32     `json:"old_locked"`
	NewLocked      int32     `json:"new_locked"`
	Reason         string    `json:"reason"`
}

// NewInventory 创建并返回一个新的 Inventory 实体实例。
func NewInventory(skuID, productID, warehouseID uint64, totalStock, warningThreshold int32) *Inventory {
	inv := &Inventory{
		SkuID:            skuID,
		ProductID:        productID,
		WarehouseID:      warehouseID,
		AvailableStock:   totalStock,
		LockedStock:      0,
		TotalStock:       totalStock,
		WarningThreshold: warningThreshold,
		PersistenceVer:   1,
	}
	// 将 uint64 ID 转换为 string 设置给 AggregateRoot
	inv.SetID(fmt.Sprintf("%d", skuID))
	inv.updateStatus()
	return inv
}

// Deduct 扣减指定数量的库存。
func (inv *Inventory) Deduct(quantity int32, reason string) (*InventoryLog, error) {
	if quantity <= 0 {
		return nil, ErrNegativeQuantity
	}
	if inv.AvailableStock < quantity {
		return nil, fmt.Errorf("%w: available=%d, required=%d", ErrInsufficientStock, inv.AvailableStock, quantity)
	}

	base := eventsourcing.NewBaseEvent(StockDeductedEventType, inv.GetID(), inv.AggregateRoot.Version())
	event := &StockDeductedEvent{
		BaseEvent: base,
		SkuID:     inv.SkuID,
		Quantity:  quantity,
		Reason:    reason,
		Timestamp: base.Timestamp,
	}
	inv.ApplyChange(event)
	inv.Apply(event)

	return inv.createLog("Deduct", -quantity, inv.AvailableStock+quantity, inv.AvailableStock, inv.LockedStock, inv.LockedStock, reason), nil
}

// Lock 锁定指定数量的库存。
func (inv *Inventory) Lock(quantity int32, reason string) (*InventoryLog, error) {
	if quantity <= 0 {
		return nil, ErrNegativeQuantity
	}
	if inv.AvailableStock < quantity {
		return nil, fmt.Errorf("%w: available=%d, required=%d", ErrInsufficientStock, inv.AvailableStock, quantity)
	}

	base := eventsourcing.NewBaseEvent(StockLockedEventType, inv.GetID(), inv.AggregateRoot.Version())
	event := &StockLockedEvent{
		BaseEvent: base,
		SkuID:     inv.SkuID,
		Quantity:  quantity,
		Reason:    reason,
		Timestamp: base.Timestamp,
	}
	inv.ApplyChange(event)
	inv.Apply(event)

	return inv.createLog("Lock", 0, inv.AvailableStock+quantity, inv.AvailableStock, inv.LockedStock-quantity, inv.LockedStock, reason), nil
}

// Unlock 解锁指定数量的库存。
func (inv *Inventory) Unlock(quantity int32, reason string) (*InventoryLog, error) {
	if quantity <= 0 {
		return nil, ErrNegativeQuantity
	}
	if inv.LockedStock < quantity {
		return nil, fmt.Errorf("%w: locked=%d, required=%d", ErrInsufficientStock, inv.LockedStock, quantity)
	}

	base := eventsourcing.NewBaseEvent(StockUnlockedEventType, inv.GetID(), inv.AggregateRoot.Version())
	event := &StockUnlockedEvent{
		BaseEvent: base,
		SkuID:     inv.SkuID,
		Quantity:  quantity,
		Reason:    reason,
		Timestamp: base.Timestamp,
	}
	inv.ApplyChange(event)
	inv.Apply(event)

	return inv.createLog("Unlock", 0, inv.AvailableStock-quantity, inv.AvailableStock, inv.LockedStock+quantity, inv.LockedStock, reason), nil
}

// ConfirmDeduction 确认扣减。
func (inv *Inventory) ConfirmDeduction(quantity int32, reason string) (*InventoryLog, error) {
	if quantity <= 0 {
		return nil, ErrNegativeQuantity
	}
	if inv.LockedStock < quantity {
		return nil, fmt.Errorf("%w: locked=%d, required=%d", ErrInsufficientStock, inv.LockedStock, quantity)
	}

	base := eventsourcing.NewBaseEvent(StockDeductedEventType, inv.GetID(), inv.AggregateRoot.Version())
	event := &StockDeductedEvent{
		BaseEvent: base,
		SkuID:     inv.SkuID,
		Quantity:  quantity,
		Reason:    reason,
		Timestamp: base.Timestamp,
	}
	inv.ApplyChange(event)
	inv.Apply(event)

	return inv.createLog("ConfirmDeduction", -quantity, inv.AvailableStock, inv.AvailableStock, inv.LockedStock+quantity, inv.LockedStock, reason), nil
}

// Add 增加库存。
func (inv *Inventory) Add(quantity int32, reason string) (*InventoryLog, error) {
	if quantity <= 0 {
		return nil, ErrNegativeQuantity
	}

	base := eventsourcing.NewBaseEvent(StockAddedEventType, inv.GetID(), inv.AggregateRoot.Version())
	event := &StockAddedEvent{
		BaseEvent: base,
		SkuID:     inv.SkuID,
		Quantity:  quantity,
		Reason:    reason,
		Timestamp: base.Timestamp,
	}
	inv.ApplyChange(event)
	inv.Apply(event)

	return inv.createLog("Add", quantity, inv.AvailableStock-quantity, inv.AvailableStock, inv.LockedStock, inv.LockedStock, reason), nil
}

// updateStatus 根据当前可用库存和总库存更新库存状态。
func (inv *Inventory) updateStatus() {
	if inv.TotalStock == 0 {
		inv.Status = InventoryStatusOutOfStock
	} else if inv.AvailableStock <= inv.WarningThreshold {
		inv.Status = InventoryStatusWarning
	} else {
		inv.Status = InventoryStatusNormal
	}
}

// createLog 创建一条库存操作日志对象。
func (inv *Inventory) createLog(action string, changeQuantity, oldAvailable, newAvailable, oldLocked, newLocked int32, reason string) *InventoryLog {
	log := &InventoryLog{
		InventoryID:    uint64(inv.ID),
		SkuID:          inv.SkuID,
		Action:         action,
		ChangeQuantity: changeQuantity,
		OldAvailable:   oldAvailable,
		NewAvailable:   newAvailable,
		OldLocked:      oldLocked,
		NewLocked:      newLocked,
		Reason:         reason,
	}
	return log
}

// Warehouse 实体代表一个仓库。
type Warehouse struct {
	ID        uint      `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Name      string    `json:"name"`
	Lat       float64   `json:"lat"`
	Lon       float64   `json:"lon"`
	Priority  int       `json:"priority"`
	ShipCost  int64     `json:"ship_cost"`
}

func NewWarehouse(name string, lat, lon float64, priority int, shipCost int64) *Warehouse {
	return &Warehouse{
		Name:     name,
		Lat:      lat,
		Lon:      lon,
		Priority: priority,
		ShipCost: shipCost,
	}
}
