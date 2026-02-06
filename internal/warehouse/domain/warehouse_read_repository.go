package domain

import "context"

// WarehouseReadRepository 定义仓库读模型的高性能访问接口（Redis）。
type WarehouseReadRepository interface {
	// Warehouse
	SaveWarehouse(ctx context.Context, warehouse *Warehouse) error
	GetWarehouse(ctx context.Context, id uint64) (*Warehouse, error)
	DeleteWarehouse(ctx context.Context, id uint64) error

	// Stock
	SaveStock(ctx context.Context, stock *WarehouseStock) error
	GetStock(ctx context.Context, warehouseID, skuID uint64) (*WarehouseStock, error)
	DeleteStock(ctx context.Context, warehouseID, skuID uint64) error

	// Transfer
	SaveTransfer(ctx context.Context, transfer *StockTransfer) error
	GetTransfer(ctx context.Context, id uint64) (*StockTransfer, error)
	DeleteTransfer(ctx context.Context, id uint64) error
}
