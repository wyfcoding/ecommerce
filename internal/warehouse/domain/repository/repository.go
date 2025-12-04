package repository

import (
	"context"
	"github.com/wyfcoding/ecommerce/internal/warehouse/domain/entity" // 导入仓库领域的实体定义。
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
	SaveWarehouse(ctx context.Context, warehouse *entity.Warehouse) error
	// GetWarehouse 根据ID获取仓库实体。
	GetWarehouse(ctx context.Context, id uint64) (*entity.Warehouse, error)
	// GetWarehouseByCode 根据仓库代码获取仓库实体。
	GetWarehouseByCode(ctx context.Context, code string) (*entity.Warehouse, error)
	// ListWarehouses 列出所有仓库实体，支持通过状态过滤和分页。
	ListWarehouses(ctx context.Context, status *entity.WarehouseStatus, offset, limit int) ([]*entity.Warehouse, int64, error)

	// --- 库存管理 (WarehouseStock methods) ---

	// SaveStock 将仓库库存实体保存到数据存储中。
	SaveStock(ctx context.Context, stock *entity.WarehouseStock) error
	// GetStock 获取指定仓库ID和SKUID的库存实体。
	GetStock(ctx context.Context, warehouseID, skuID uint64) (*entity.WarehouseStock, error)
	// ListStocks 列出指定仓库ID的所有库存实体，支持分页。
	ListStocks(ctx context.Context, warehouseID uint64, offset, limit int) ([]*entity.WarehouseStock, int64, error)

	// --- 调拨管理 (StockTransfer methods) ---

	// SaveTransfer 将库存调拨实体保存到数据存储中。
	SaveTransfer(ctx context.Context, transfer *entity.StockTransfer) error
	// GetTransfer 根据ID获取库存调拨实体。
	GetTransfer(ctx context.Context, id uint64) (*entity.StockTransfer, error)
	// ListTransfers 列出指定调出/调入仓库ID和状态的所有库存调拨实体，支持分页。
	ListTransfers(ctx context.Context, fromWarehouseID, toWarehouseID uint64, status *entity.StockTransferStatus, offset, limit int) ([]*entity.StockTransfer, int64, error)
}
