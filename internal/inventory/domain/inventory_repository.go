package domain

import (
	"context"
)

// InventoryRepository 是库存模块的仓储接口。
// 它定义了对库存实体和库存日志实体进行数据持久化操作的契约。
// 仓储接口属于领域层，旨在将领域逻辑与数据存储的实现细节解耦。
type InventoryRepository interface {
	// Save 将库存实体保存到数据存储中。
	// 如果库存已存在，则更新；如果不存在，则创建。
	// ctx: 上下文。
	// inventory: 待保存的库存实体。
	Save(ctx context.Context, inventory *Inventory) error
	// GetBySkuID 根据SKU ID获取库存实体。
	GetBySkuID(ctx context.Context, skuID uint64) (*Inventory, error)
	// GetBySkuIDs 根据SKU ID列表获取多个库存实体。
	GetBySkuIDs(ctx context.Context, skuIDs []uint64) ([]*Inventory, error)
	// List 列出所有库存实体，支持分页。
	List(ctx context.Context, offset, limit int) ([]*Inventory, int64, error)
	// GetLogs 获取指定库存ID的所有库存日志。
	GetLogs(ctx context.Context, inventoryID uint64, offset, limit int) ([]*InventoryLog, int64, error)
}

// WarehouseRepository 是仓库模块的仓储接口。
// 它定义了对仓库实体进行数据持久化操作的契约。
// 仓储接口属于领域层。
type WarehouseRepository interface {
	// Save 将仓库实体保存到数据存储中。
	Save(ctx context.Context, warehouse *Warehouse) error
	// GetByID 根据ID获取仓库实体。
	GetByID(ctx context.Context, id uint64) (*Warehouse, error)
	// ListAll 列出所有仓库实体。
	ListAll(ctx context.Context) ([]*Warehouse, error)
}
