package persistence

import (
	"fmt"

	"github.com/wyfcoding/ecommerce/internal/inventory/domain"
	"gorm.io/gorm"
)

// InventoryModel 库存写模型（持久化专用）。
type InventoryModel struct {
	gorm.Model
	SkuID            uint64                 `gorm:"column:sku_id;not null;index;comment:SKU ID"`
	ProductID        uint64                 `gorm:"column:product_id;not null;index;comment:商品ID"`
	WarehouseID      uint64                 `gorm:"column:warehouse_id;not null;index;comment:仓库ID"`
	AvailableStock   int32                  `gorm:"column:available_stock;not null;default:0;comment:可用库存"`
	LockedStock      int32                  `gorm:"column:locked_stock;not null;default:0;comment:锁定库存"`
	TotalStock       int32                  `gorm:"column:total_stock;not null;default:0;comment:总库存"`
	Status           domain.InventoryStatus `gorm:"column:status;default:1;comment:状态"`
	WarningThreshold int32                  `gorm:"column:warning_threshold;default:10;comment:预警阈值"`
	Version          int64                  `gorm:"column:version;default:1;comment:乐观锁版本号"`
}

func (InventoryModel) TableName() string {
	return "inventories"
}

// InventoryLogModel 库存日志写模型（持久化专用）。
type InventoryLogModel struct {
	gorm.Model
	InventoryID    uint64 `gorm:"column:inventory_id;not null;index;comment:库存ID"`
	SkuID          uint64 `gorm:"column:sku_id;not null;index;comment:SKU ID"`
	Action         string `gorm:"column:action;type:varchar(32);not null;comment:操作类型"`
	ChangeQuantity int32  `gorm:"column:change_quantity;not null;comment:变更数量"`
	OldAvailable   int32  `gorm:"column:old_available;not null;comment:变更前可用"`
	NewAvailable   int32  `gorm:"column:new_available;not null;comment:变更后可用"`
	OldLocked      int32  `gorm:"column:old_locked;not null;comment:变更前锁定"`
	NewLocked      int32  `gorm:"column:new_locked;not null;comment:变更后锁定"`
	Reason         string `gorm:"column:reason;type:varchar(255);comment:原因"`
}

func (InventoryLogModel) TableName() string {
	return "inventory_logs"
}

// WarehouseModel 仓库写模型（持久化专用）。
type WarehouseModel struct {
	gorm.Model
	Name     string  `gorm:"column:name;type:varchar(255);not null;comment:仓库名称"`
	Lat      float64 `gorm:"column:lat;type:decimal(10,6);not null;comment:纬度"`
	Lon      float64 `gorm:"column:lon;type:decimal(10,6);not null;comment:经度"`
	Priority int     `gorm:"column:priority;not null;default:0;comment:优先级"`
	ShipCost int64   `gorm:"column:ship_cost;not null;default:0;comment:基础配送成本(分)"`
}

func (WarehouseModel) TableName() string {
	return "warehouses"
}

func toInventoryModel(inv *domain.Inventory) *InventoryModel {
	if inv == nil {
		return nil
	}
	return &InventoryModel{
		Model: gorm.Model{
			ID:        inv.ID,
			CreatedAt: inv.CreatedAt,
			UpdatedAt: inv.UpdatedAt,
		},
		SkuID:            inv.SkuID,
		ProductID:        inv.ProductID,
		WarehouseID:      inv.WarehouseID,
		AvailableStock:   inv.AvailableStock,
		LockedStock:      inv.LockedStock,
		TotalStock:       inv.TotalStock,
		Status:           inv.Status,
		WarningThreshold: inv.WarningThreshold,
		Version:          inv.PersistenceVer,
	}
}

func toDomainInventory(model *InventoryModel) *domain.Inventory {
	if model == nil {
		return nil
	}
	inv := &domain.Inventory{
		ID:               model.ID,
		CreatedAt:        model.CreatedAt,
		UpdatedAt:        model.UpdatedAt,
		SkuID:            model.SkuID,
		ProductID:        model.ProductID,
		WarehouseID:      model.WarehouseID,
		AvailableStock:   model.AvailableStock,
		LockedStock:      model.LockedStock,
		TotalStock:       model.TotalStock,
		Status:           model.Status,
		WarningThreshold: model.WarningThreshold,
		PersistenceVer:   model.Version,
	}
	inv.SetID(fmt.Sprintf("%d", inv.SkuID))
	inv.SetVersion(inv.PersistenceVer)
	return inv
}

func toInventoryLogModel(log *domain.InventoryLog) *InventoryLogModel {
	if log == nil {
		return nil
	}
	return &InventoryLogModel{
		Model: gorm.Model{
			ID:        log.ID,
			CreatedAt: log.CreatedAt,
			UpdatedAt: log.UpdatedAt,
		},
		InventoryID:    log.InventoryID,
		SkuID:          log.SkuID,
		Action:         log.Action,
		ChangeQuantity: log.ChangeQuantity,
		OldAvailable:   log.OldAvailable,
		NewAvailable:   log.NewAvailable,
		OldLocked:      log.OldLocked,
		NewLocked:      log.NewLocked,
		Reason:         log.Reason,
	}
}

func toDomainInventoryLog(model *InventoryLogModel) *domain.InventoryLog {
	if model == nil {
		return nil
	}
	return &domain.InventoryLog{
		ID:             model.ID,
		CreatedAt:      model.CreatedAt,
		UpdatedAt:      model.UpdatedAt,
		InventoryID:    model.InventoryID,
		SkuID:          model.SkuID,
		Action:         model.Action,
		ChangeQuantity: model.ChangeQuantity,
		OldAvailable:   model.OldAvailable,
		NewAvailable:   model.NewAvailable,
		OldLocked:      model.OldLocked,
		NewLocked:      model.NewLocked,
		Reason:         model.Reason,
	}
}

func toWarehouseModel(warehouse *domain.Warehouse) *WarehouseModel {
	if warehouse == nil {
		return nil
	}
	return &WarehouseModel{
		Model: gorm.Model{
			ID:        warehouse.ID,
			CreatedAt: warehouse.CreatedAt,
			UpdatedAt: warehouse.UpdatedAt,
		},
		Name:     warehouse.Name,
		Lat:      warehouse.Lat,
		Lon:      warehouse.Lon,
		Priority: warehouse.Priority,
		ShipCost: warehouse.ShipCost,
	}
}

func toDomainWarehouse(model *WarehouseModel) *domain.Warehouse {
	if model == nil {
		return nil
	}
	return &domain.Warehouse{
		ID:        model.ID,
		CreatedAt: model.CreatedAt,
		UpdatedAt: model.UpdatedAt,
		Name:      model.Name,
		Lat:       model.Lat,
		Lon:       model.Lon,
		Priority:  model.Priority,
		ShipCost:  model.ShipCost,
	}
}
