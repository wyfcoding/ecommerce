package data

import (
	"gorm.io/gorm"
)

// Stock represents the quantity of a SKU in a specific warehouse.
type Stock struct {
	gorm.Model
	SKUID          uint64 `gorm:"uniqueIndex:idx_sku_warehouse;not null;comment:SKU ID" json:"skuId"`
	WarehouseID    uint64 `gorm:"uniqueIndex:idx_sku_warehouse;not null;comment:仓库ID" json:"warehouseId"`
	Quantity       uint32 `gorm:"not null;default:0;comment:库存数量" json:"quantity"`
	LockedQuantity uint32 `gorm:"not null;default:0;comment:锁定库存数量" json:"lockedQuantity"` // For pre-allocation
}

// Warehouse represents a physical warehouse.
type Warehouse struct {
	gorm.Model
	Name     string `gorm:"size:255;not null;unique;comment:仓库名称" json:"name"`
	Location string `gorm:"size:512;comment:仓库地址" json:"location"`
}

// TableName specifies the table name for Stock.
func (Stock) TableName() string {
	return "stocks"
}

// TableName specifies the table name for Warehouse.
func (Warehouse) TableName() string {
	return "warehouses"
}
