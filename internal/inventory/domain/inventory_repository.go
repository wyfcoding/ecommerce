package domain

import (
	"context"

	"github.com/wyfcoding/pkg/eventsourcing"
)

// EventStore 库存模块的事件存储接口。
type EventStore interface {
	// Save 保存领域事件。
	Save(ctx context.Context, events []eventsourcing.DomainEvent) error
	// Load 获取聚合根的事件历史（阶段B验收主接口）。
	Load(ctx context.Context, aggregateID string) ([]eventsourcing.DomainEvent, error)
	// GetHistory 获取聚合根的事件历史。
	GetHistory(ctx context.Context, aggregateID string) ([]eventsourcing.DomainEvent, error)
}

// InventoryRepository 是库存模块的仓储接口。
type InventoryRepository interface {
	Save(ctx context.Context, inventory *Inventory) error
	SaveWithOptimisticLock(ctx context.Context, inventory *Inventory) error
	SaveLog(ctx context.Context, log *InventoryLog) error
	GetBySkuID(ctx context.Context, skuID uint64) (*Inventory, error)
	GetBySkuIDs(ctx context.Context, skuIDs []uint64) ([]*Inventory, error)
	GetInventoryByWarehouse(ctx context.Context, warehouseID string) ([]*InventoryItem, error)
	List(ctx context.Context, offset, limit int) ([]*Inventory, int64, error)
	GetLogs(ctx context.Context, skuID uint64, inventoryID uint64, offset, limit int) ([]*InventoryLog, int64, error)
	Delete(ctx context.Context, skuID uint64) error
	// ExecWithBarrier 在分布式事务屏障下执行业务逻辑
	ExecWithBarrier(ctx context.Context, barrier any, fn func(ctx context.Context) error) error
	Reserve(ctx context.Context, skuID uint64, quantity int32) error
}

// WarehouseRepository 是仓库模块的仓储接口。
// WarehouseRepository 是仓库模块的仓储接口。
type WarehouseRepository interface {
	SaveWarehouse(ctx context.Context, warehouse *Warehouse) error
	GetWarehouse(ctx context.Context, warehouseID string) (*Warehouse, error)
	GetWarehouseByCode(ctx context.Context, warehouseCode string) (*Warehouse, error)
	GetActiveWarehouses(ctx context.Context) ([]*Warehouse, error)
	GetWarehousesByType(ctx context.Context, warehouseType WarehouseType) ([]*Warehouse, error)
	UpdateWarehouse(ctx context.Context, warehouse *Warehouse) error
	DeleteWarehouse(ctx context.Context, warehouseID string) error

	// Legacy support or alias if needed
	Save(ctx context.Context, warehouse *Warehouse) error
	GetByID(ctx context.Context, id uint64) (*Warehouse, error)
	ListAll(ctx context.Context) ([]*Warehouse, error)
}
