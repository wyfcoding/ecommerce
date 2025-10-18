package model

import (
	"time"

	"gorm.io/gorm"
)

// MovementType 定义了库存流动的类型
type MovementType string

const (
	MovementTypeInbound    MovementType = "INBOUND"    // 入库 (采购、退货入库)
	MovementTypeOutbound   MovementType = "OUTBOUND"   // 出库 (销售)
	MovementTypeAdjustment MovementType = "ADJUSTMENT" // 调整 (盘点)
	MovementTypeRollback   MovementType = "ROLLBACK"   // 回滚 (订单取消)
)

// Inventory 库存主模型
// 记录了每个 SKU 的核心库存信息
type Inventory struct {
	ID          uint `gorm:"primarykey" json:"id"`
	ProductSKU  string `gorm:"type:varchar(100);uniqueIndex;not null" json:"product_sku"` // 关联的商品 SKU
	WarehouseID uint `gorm:"not null;index" json:"warehouse_id"`                      // 仓库ID

	// 核心库存字段
	QuantityOnHand int `gorm:"not null;default:0" json:"quantity_on_hand"` // 现有库存 (物理库存)
	QuantityReserved int `gorm:"not null;default:0" json:"quantity_reserved"` // 预留库存 (已被下单但未发货)
	QuantityAvailable int `gorm:"not null;default:0" json:"quantity_available"` // 可用库存 (On-Hand - Reserved)

	UpdatedAt time.Time `json:"updated_at"`
}

// StockMovement 库存流水模型
// 记录每一次库存变动，用于追踪和审计
type StockMovement struct {
	ID           uint         `gorm:"primarykey" json:"id"`
	InventoryID  uint         `gorm:"not null;index" json:"inventory_id"` // 关联的库存记录ID
	Type         MovementType `gorm:"type:varchar(20);not null" json:"type"`  // 流动类型
	Quantity     int          `gorm:"not null" json:"quantity"`              // 变动数量 (正数表示增加, 负数表示减少)
	Reference    string       `gorm:"type:varchar(100);index" json:"reference"` // 关联单号 (如订单号、采购单号)
	Reason       string       `gorm:com:"type:varchar(255)" json:"reason"`    // 变动原因
	CreatedAt    time.Time    `json:"created_at"`
}

// Warehouse 仓库模型 (简化)
type Warehouse struct {
	ID      uint   `gorm:"primarykey" json:"id"`
	Name    string `gorm:"type:varchar(100);not null" json:"name"`
	Address string `gorm:"type:varchar(255)" json:"address"`
}

// TableName 自定义表名
func (Inventory) TableName() string {
	return "inventories"
}

func (StockMovement) TableName() string {
	return "stock_movements"
}

func (Warehouse) TableName() string {
	return "warehouses"
}
