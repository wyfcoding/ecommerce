package repository

import (
	"context"
	"ecommerce/internal/warehouse/domain/entity"
)

// WarehouseRepository 仓库仓储接口
type WarehouseRepository interface {
	// 仓库管理
	SaveWarehouse(ctx context.Context, warehouse *entity.Warehouse) error
	GetWarehouse(ctx context.Context, id uint64) (*entity.Warehouse, error)
	GetWarehouseByCode(ctx context.Context, code string) (*entity.Warehouse, error)
	ListWarehouses(ctx context.Context, status *entity.WarehouseStatus, offset, limit int) ([]*entity.Warehouse, int64, error)

	// 库存管理
	SaveStock(ctx context.Context, stock *entity.WarehouseStock) error
	GetStock(ctx context.Context, warehouseID, skuID uint64) (*entity.WarehouseStock, error)
	ListStocks(ctx context.Context, warehouseID uint64, offset, limit int) ([]*entity.WarehouseStock, int64, error)

	// 调拨管理
	SaveTransfer(ctx context.Context, transfer *entity.StockTransfer) error
	GetTransfer(ctx context.Context, id uint64) (*entity.StockTransfer, error)
	ListTransfers(ctx context.Context, fromWarehouseID, toWarehouseID uint64, status *entity.StockTransferStatus, offset, limit int) ([]*entity.StockTransfer, int64, error)
}
