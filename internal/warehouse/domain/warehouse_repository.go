package domain

import (
	"context"
)

// WarehouseRepository 是仓库模块的仓储接口。
// 它定义了对仓库、仓库库存和库存调拨实体进行数据持久化操作的契约。
// 仓储接口属于领域层，旨在将领域逻辑与数据存储的实现细节解耦。
type WarehouseRepository interface {
	// --- 仓库管理 (Warehouse methods) ---

	// SaveWarehouse 将仓库实体保存到数据存储中。
	// 如果实体已存在，则更新；如果不存在，则创建。
	// ctx: 上下文。
	// warehouse: 待保存的仓库实体。
	SaveWarehouse(ctx context.Context, warehouse *Warehouse) error
	// GetWarehouse 根据ID获取仓库实体。
	GetWarehouse(ctx context.Context, id uint64) (*Warehouse, error)
	// GetWarehouseByCode 根据仓库代码获取仓库实体。
	GetWarehouseByCode(ctx context.Context, code string) (*Warehouse, error)
	// ListWarehouses 列出所有仓库实体，支持通过状态过滤和分页。
	ListWarehouses(ctx context.Context, status *WarehouseStatus, offset, limit int) ([]*Warehouse, int64, error)
	// ListWarehousesWithStock 获取拥有指定 SKU 且可用库存足够的仓库及其库存量。
	ListWarehousesWithStock(ctx context.Context, skuID uint64, minQty int32) ([]*Warehouse, []int32, error)

	// --- 库存管理 (WarehouseStock methods) ---

	// SaveStock 将仓库库存实体保存到数据存储中。
	SaveStock(ctx context.Context, stock *WarehouseStock) error
	// GetStock 获取指定仓库ID和SKUID的库存实体。
	GetStock(ctx context.Context, warehouseID, skuID uint64) (*WarehouseStock, error)
	// ListStocks 列出指定仓库ID的所有库存实体，支持分页。
	ListStocks(ctx context.Context, warehouseID uint64, offset, limit int) ([]*WarehouseStock, int64, error)

	// --- 调拨管理 (StockTransfer methods) ---

	// SaveTransfer 将库存调拨实体保存到数据存储中。
	SaveTransfer(ctx context.Context, transfer *StockTransfer) error
	// GetTransfer 根据ID获取库存调拨实体。
	GetTransfer(ctx context.Context, id uint64) (*StockTransfer, error)
	// ListTransfers 列出指定调出/调入仓库ID和状态的所有库存调拨实体，支持分页。
	ListTransfers(ctx context.Context, fromWarehouseID, toWarehouseID uint64, status *StockTransferStatus, offset, limit int) ([]*StockTransfer, int64, error)
}
