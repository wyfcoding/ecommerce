package entity

import (
	"errors" // 导入标准错误处理库。
	"fmt"    // 导入格式化库。

	"gorm.io/gorm" // 导入GORM库。
)

// 定义Inventory模块的业务错误。
var (
	ErrInsufficientStock = errors.New("库存不足")    // 库存不足以完成操作。
	ErrNegativeQuantity  = errors.New("数量必须为正数") // 操作数量必须为正数。
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
// 它代表一个SKU在特定仓库中的库存信息，包含了可用库存、锁定库存、总库存和状态。
type Inventory struct {
	gorm.Model                       // 嵌入gorm.Model，包含ID, CreatedAt, UpdatedAt, DeletedAt等通用字段。
	SkuID            uint64          `gorm:"not null;index;comment:SKU ID" json:"sku_id"`            // 关联的SKU ID，索引字段。
	ProductID        uint64          `gorm:"not null;index;comment:商品ID" json:"product_id"`          // 关联的商品ID，索引字段。
	WarehouseID      uint64          `gorm:"not null;index;comment:仓库ID" json:"warehouse_id"`        // 关联的仓库ID，索引字段。
	AvailableStock   int32           `gorm:"not null;default:0;comment:可用库存" json:"available_stock"` // 可用于销售的库存数量。
	LockedStock      int32           `gorm:"not null;default:0;comment:锁定库存" json:"locked_stock"`    // 因预购或订单待支付而锁定的库存数量。
	TotalStock       int32           `gorm:"not null;default:0;comment:总库存" json:"total_stock"`      // 总库存数量（可用库存 + 锁定库存）。
	Status           InventoryStatus `gorm:"default:1;comment:状态" json:"status"`                     // 库存状态，默认为正常。
	WarningThreshold int32           `gorm:"default:10;comment:预警阈值" json:"warning_threshold"`       // 触发库存预警的阈值。
	Logs             []*InventoryLog `gorm:"foreignKey:InventoryID" json:"logs"`                     // 关联的库存操作日志列表，一对多关系。
}

// InventoryLog 实体代表库存的一次操作日志。
// 它记录了操作类型、数量变更、变更前后状态和原因等信息。
type InventoryLog struct {
	gorm.Model            // 嵌入gorm.Model。
	InventoryID    uint64 `gorm:"not null;index;comment:库存ID" json:"inventory_id"`      // 关联的库存记录ID，索引字段。
	Action         string `gorm:"type:varchar(32);not null;comment:操作类型" json:"action"` // 操作类型，例如“Add”（增加），“Deduct”（扣减），“Lock”（锁定）。
	ChangeQuantity int32  `gorm:"not null;comment:变更数量" json:"change_quantity"`         // 本次操作导致的库存数量变化。
	OldAvailable   int32  `gorm:"not null;comment:变更前可用" json:"old_available"`          // 变更前的可用库存数量。
	NewAvailable   int32  `gorm:"not null;comment:变更后可用" json:"new_available"`          // 变更后的可用库存数量。
	OldLocked      int32  `gorm:"not null;comment:变更前锁定" json:"old_locked"`             // 变更前的锁定库存数量。
	NewLocked      int32  `gorm:"not null;comment:变更后锁定" json:"new_locked"`             // 变更后的锁定库存数量。
	Reason         string `gorm:"type:varchar(255);comment:原因" json:"reason"`           // 变更原因。
}

// NewInventory 创建并返回一个新的 Inventory 实体实例。
// skuID, productID, warehouseID: SKU、商品和仓库ID。
// totalStock: 初始总库存。
// warningThreshold: 预警阈值。
func NewInventory(skuID, productID, warehouseID uint64, totalStock, warningThreshold int32) *Inventory {
	status := InventoryStatusNormal
	if totalStock == 0 {
		status = InventoryStatusOutOfStock
	} else if totalStock <= warningThreshold {
		status = InventoryStatusWarning
	}

	return &Inventory{
		SkuID:            skuID,
		ProductID:        productID,
		WarehouseID:      warehouseID,
		AvailableStock:   totalStock, // 初始可用库存等于总库存。
		LockedStock:      0,          // 初始锁定库存为0。
		TotalStock:       totalStock,
		Status:           status,
		WarningThreshold: warningThreshold,
	}
}

// CanDeduct 检查是否可以扣减指定数量的库存。
// quantity: 待扣减的数量。
func (inv *Inventory) CanDeduct(quantity int32) error {
	if quantity <= 0 {
		return ErrNegativeQuantity
	}
	if inv.AvailableStock < quantity {
		// 返回详细错误信息，指明可用库存和所需数量。
		return fmt.Errorf("%w: available=%d, required=%d", ErrInsufficientStock, inv.AvailableStock, quantity)
	}
	return nil
}

// Deduct 扣减指定数量的库存。
// quantity: 待扣减的数量。
// reason: 扣减原因。
func (inv *Inventory) Deduct(quantity int32, reason string) error {
	if err := inv.CanDeduct(quantity); err != nil {
		return err
	}

	oldAvailable := inv.AvailableStock
	// 扣减可用库存和总库存。
	inv.AvailableStock -= quantity
	inv.TotalStock -= quantity

	inv.updateStatus() // 更新库存状态。
	// 记录库存日志。
	inv.AddLog("Deduct", -quantity, oldAvailable, inv.AvailableStock, inv.LockedStock, inv.LockedStock, reason)

	return nil
}

// CanLock 检查是否可以锁定指定数量的库存。
// quantity: 待锁定的数量。
func (inv *Inventory) CanLock(quantity int32) error {
	if quantity <= 0 {
		return ErrNegativeQuantity
	}
	if inv.AvailableStock < quantity {
		// 返回详细错误信息，指明可用库存和所需数量。
		return fmt.Errorf("%w: available=%d, required=%d", ErrInsufficientStock, inv.AvailableStock, quantity)
	}
	return nil
}

// Lock 锁定指定数量的库存。
// quantity: 待锁定的数量。
// reason: 锁定原因。
func (inv *Inventory) Lock(quantity int32, reason string) error {
	if err := inv.CanLock(quantity); err != nil {
		return err
	}

	oldAvailable := inv.AvailableStock
	oldLocked := inv.LockedStock

	// 减少可用库存，增加锁定库存。
	inv.AvailableStock -= quantity
	inv.LockedStock += quantity

	inv.updateStatus() // 更新库存状态。
	// 记录库存日志。
	inv.AddLog("Lock", 0, oldAvailable, inv.AvailableStock, oldLocked, inv.LockedStock, reason) // 数量变更设为0，表示只是状态转移。

	return nil
}

// CanUnlock 检查是否可以解锁指定数量的库存。
// quantity: 待解锁的数量。
func (inv *Inventory) CanUnlock(quantity int32) error {
	if quantity <= 0 {
		return ErrNegativeQuantity
	}
	if inv.LockedStock < quantity {
		// 返回详细错误信息，指明锁定库存和所需数量。
		return fmt.Errorf("%w: locked=%d, required=%d", ErrInsufficientStock, inv.LockedStock, quantity)
	}
	return nil
}

// Unlock 解锁指定数量的库存。
// quantity: 待解锁的数量。
// reason: 解锁原因。
func (inv *Inventory) Unlock(quantity int32, reason string) error {
	if err := inv.CanUnlock(quantity); err != nil {
		return err
	}

	oldAvailable := inv.AvailableStock
	oldLocked := inv.LockedStock

	// 增加可用库存，减少锁定库存。
	inv.AvailableStock += quantity
	inv.LockedStock -= quantity

	inv.updateStatus() // 更新库存状态。
	// 记录库存日志。
	inv.AddLog("Unlock", 0, oldAvailable, inv.AvailableStock, oldLocked, inv.LockedStock, reason) // 数量变更设为0，表示只是状态转移。

	return nil
}

// CanConfirmDeduction 检查是否可以确认扣减指定数量的库存（从锁定库存中扣减）。
// quantity: 待确认扣减的数量。
func (inv *Inventory) CanConfirmDeduction(quantity int32) error {
	if quantity <= 0 {
		return ErrNegativeQuantity
	}
	if inv.LockedStock < quantity {
		// 返回详细错误信息，指明锁定库存和所需数量。
		return fmt.Errorf("%w: locked=%d, required=%d", ErrInsufficientStock, inv.LockedStock, quantity)
	}
	return nil
}

// ConfirmDeduction 确认扣减指定数量的库存（从锁定库存中扣减）。
// quantity: 待确认扣减的数量。
// reason: 确认扣减原因。
func (inv *Inventory) ConfirmDeduction(quantity int32, reason string) error {
	if err := inv.CanConfirmDeduction(quantity); err != nil {
		return err
	}

	oldLocked := inv.LockedStock

	// 减少锁定库存和总库存。
	inv.LockedStock -= quantity
	inv.TotalStock -= quantity

	inv.updateStatus() // 更新库存状态。
	// 记录库存日志。
	inv.AddLog("ConfirmDeduction", -quantity, inv.AvailableStock, inv.AvailableStock, oldLocked, inv.LockedStock, reason)

	return nil
}

// CanAdd 检查是否可以增加指定数量的库存。
// quantity: 待增加的数量。
func (inv *Inventory) CanAdd(quantity int32) error {
	if quantity <= 0 {
		return ErrNegativeQuantity
	}
	return nil
}

// Add 增加指定数量的库存。
// quantity: 待增加的数量。
// reason: 增加库存的原因。
func (inv *Inventory) Add(quantity int32, reason string) error {
	if err := inv.CanAdd(quantity); err != nil {
		return err
	}

	oldAvailable := inv.AvailableStock
	// 增加可用库存和总库存。
	inv.AvailableStock += quantity
	inv.TotalStock += quantity

	inv.updateStatus() // 更新库存状态。
	// 记录库存日志。
	inv.AddLog("Add", quantity, oldAvailable, inv.AvailableStock, inv.LockedStock, inv.LockedStock, reason)

	return nil
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

// AddLog 添加一条库存操作日志到当前库存记录。
// action: 操作类型。
// changeQuantity: 本次操作导致的库存数量变化。
// oldAvailable, newAvailable: 变更前后可用库存。
// oldLocked, newLocked: 变更前后锁定库存。
// reason: 操作原因。
func (inv *Inventory) AddLog(action string, changeQuantity, oldAvailable, newAvailable, oldLocked, newLocked int32, reason string) {
	log := &InventoryLog{
		Action:         action,
		ChangeQuantity: changeQuantity,
		OldAvailable:   oldAvailable,
		NewAvailable:   newAvailable,
		OldLocked:      oldLocked,
		NewLocked:      newLocked,
		Reason:         reason,
	}
	inv.Logs = append(inv.Logs, log) // 将日志添加到关联的Logs切片。
}
