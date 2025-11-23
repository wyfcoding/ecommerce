package entity

import (
	"errors"
	"fmt"

	"gorm.io/gorm"
)

var (
	ErrInsufficientStock = errors.New("insufficient stock")
	ErrNegativeQuantity  = errors.New("quantity must be positive")
)

// InventoryStatus 库存状态
type InventoryStatus int

const (
	InventoryStatusNormal     InventoryStatus = 1 // 正常
	InventoryStatusLocked     InventoryStatus = 2 // 已锁定
	InventoryStatusWarning    InventoryStatus = 3 // 预警
	InventoryStatusOutOfStock InventoryStatus = 4 // 缺货
)

// Inventory 库存聚合根
type Inventory struct {
	gorm.Model
	SkuID            uint64          `gorm:"not null;uniqueIndex;comment:SKU ID" json:"sku_id"`
	ProductID        uint64          `gorm:"not null;index;comment:商品ID" json:"product_id"`
	AvailableStock   int32           `gorm:"not null;default:0;comment:可用库存" json:"available_stock"`
	LockedStock      int32           `gorm:"not null;default:0;comment:锁定库存" json:"locked_stock"`
	TotalStock       int32           `gorm:"not null;default:0;comment:总库存" json:"total_stock"`
	Status           InventoryStatus `gorm:"default:1;comment:状态" json:"status"`
	WarningThreshold int32           `gorm:"default:10;comment:预警阈值" json:"warning_threshold"`
	Logs             []*InventoryLog `gorm:"foreignKey:InventoryID" json:"logs"`
}

// InventoryLog 库存日志值对象
type InventoryLog struct {
	gorm.Model
	InventoryID    uint64 `gorm:"not null;index;comment:库存ID" json:"inventory_id"`
	Action         string `gorm:"type:varchar(32);not null;comment:操作类型" json:"action"`
	ChangeQuantity int32  `gorm:"not null;comment:变更数量" json:"change_quantity"`
	OldAvailable   int32  `gorm:"not null;comment:变更前可用" json:"old_available"`
	NewAvailable   int32  `gorm:"not null;comment:变更后可用" json:"new_available"`
	OldLocked      int32  `gorm:"not null;comment:变更前锁定" json:"old_locked"`
	NewLocked      int32  `gorm:"not null;comment:变更后锁定" json:"new_locked"`
	Reason         string `gorm:"type:varchar(255);comment:原因" json:"reason"`
}

// NewInventory 创建库存
func NewInventory(skuID, productID uint64, totalStock, warningThreshold int32) *Inventory {
	status := InventoryStatusNormal
	if totalStock == 0 {
		status = InventoryStatusOutOfStock
	} else if totalStock <= warningThreshold {
		status = InventoryStatusWarning
	}

	return &Inventory{
		SkuID:            skuID,
		ProductID:        productID,
		AvailableStock:   totalStock,
		LockedStock:      0,
		TotalStock:       totalStock,
		Status:           status,
		WarningThreshold: warningThreshold,
	}
}

// CanDeduct 是否可以扣减库存
func (inv *Inventory) CanDeduct(quantity int32) error {
	if quantity <= 0 {
		return ErrNegativeQuantity
	}
	if inv.AvailableStock < quantity {
		return fmt.Errorf("%w: available=%d, required=%d", ErrInsufficientStock, inv.AvailableStock, quantity)
	}
	return nil
}

// Deduct 扣减库存
func (inv *Inventory) Deduct(quantity int32, reason string) error {
	if err := inv.CanDeduct(quantity); err != nil {
		return err
	}

	oldAvailable := inv.AvailableStock
	inv.AvailableStock -= quantity
	inv.TotalStock -= quantity

	inv.updateStatus()
	inv.AddLog("Deduct", -quantity, oldAvailable, inv.AvailableStock, 0, 0, reason)

	return nil
}

// CanLock 是否可以锁定库存
func (inv *Inventory) CanLock(quantity int32) error {
	if quantity <= 0 {
		return ErrNegativeQuantity
	}
	if inv.AvailableStock < quantity {
		return fmt.Errorf("%w: available=%d, required=%d", ErrInsufficientStock, inv.AvailableStock, quantity)
	}
	return nil
}

// Lock 锁定库存
func (inv *Inventory) Lock(quantity int32, reason string) error {
	if err := inv.CanLock(quantity); err != nil {
		return err
	}

	oldAvailable := inv.AvailableStock
	oldLocked := inv.LockedStock

	inv.AvailableStock -= quantity
	inv.LockedStock += quantity

	inv.updateStatus()
	inv.AddLog("Lock", 0, oldAvailable, inv.AvailableStock, oldLocked, inv.LockedStock, reason)

	return nil
}

// CanUnlock 是否可以解锁库存
func (inv *Inventory) CanUnlock(quantity int32) error {
	if quantity <= 0 {
		return ErrNegativeQuantity
	}
	if inv.LockedStock < quantity {
		return fmt.Errorf("%w: locked=%d, required=%d", ErrInsufficientStock, inv.LockedStock, quantity)
	}
	return nil
}

// Unlock 解锁库存
func (inv *Inventory) Unlock(quantity int32, reason string) error {
	if err := inv.CanUnlock(quantity); err != nil {
		return err
	}

	oldAvailable := inv.AvailableStock
	oldLocked := inv.LockedStock

	inv.AvailableStock += quantity
	inv.LockedStock -= quantity

	inv.updateStatus()
	inv.AddLog("Unlock", 0, oldAvailable, inv.AvailableStock, oldLocked, inv.LockedStock, reason)

	return nil
}

// CanConfirmDeduction 是否可以确认扣减
func (inv *Inventory) CanConfirmDeduction(quantity int32) error {
	if quantity <= 0 {
		return ErrNegativeQuantity
	}
	if inv.LockedStock < quantity {
		return fmt.Errorf("%w: locked=%d, required=%d", ErrInsufficientStock, inv.LockedStock, quantity)
	}
	return nil
}

// ConfirmDeduction 确认扣减（从锁定库存中扣减）
func (inv *Inventory) ConfirmDeduction(quantity int32, reason string) error {
	if err := inv.CanConfirmDeduction(quantity); err != nil {
		return err
	}

	oldLocked := inv.LockedStock

	inv.LockedStock -= quantity
	inv.TotalStock -= quantity

	inv.updateStatus()
	inv.AddLog("ConfirmDeduction", -quantity, inv.AvailableStock, inv.AvailableStock, oldLocked, inv.LockedStock, reason)

	return nil
}

// CanAdd 是否可以增加库存
func (inv *Inventory) CanAdd(quantity int32) error {
	if quantity <= 0 {
		return ErrNegativeQuantity
	}
	return nil
}

// Add 增加库存
func (inv *Inventory) Add(quantity int32, reason string) error {
	if err := inv.CanAdd(quantity); err != nil {
		return err
	}

	oldAvailable := inv.AvailableStock
	inv.AvailableStock += quantity
	inv.TotalStock += quantity

	inv.updateStatus()
	inv.AddLog("Add", quantity, oldAvailable, inv.AvailableStock, 0, 0, reason)

	return nil
}

// updateStatus 更新库存状态
func (inv *Inventory) updateStatus() {
	if inv.TotalStock == 0 {
		inv.Status = InventoryStatusOutOfStock
	} else if inv.AvailableStock <= inv.WarningThreshold {
		inv.Status = InventoryStatusWarning
	} else {
		inv.Status = InventoryStatusNormal
	}
}

// AddLog 添加日志
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
	inv.Logs = append(inv.Logs, log)
}
