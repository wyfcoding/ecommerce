package domain

import (
	"context"
)

// WarehouseRepository 定义了仓库及库存数据的持久化接口。
type WarehouseRepository interface {
	// 基础 CRUD
	Save(ctx context.Context, warehouse *Warehouse) error
	FindByID(ctx context.Context, id uint64) (*Warehouse, error)
	FindByCode(ctx context.Context, code string) (*Warehouse, error)
	List(ctx context.Context, offset, limit int) ([]*Warehouse, int64, error)

	// 库存操作
	GetStock(ctx context.Context, warehouseID, skuID uint64) (*WarehouseStock, error)
	SaveStock(ctx context.Context, stock *WarehouseStock) error
	UpdateStock(ctx context.Context, stock *WarehouseStock) error

	// 库存操作 (事务内/带锁)
	GetStockWithLock(ctx context.Context, tx any, warehouseID, skuID uint64) (*WarehouseStock, error)
	SaveStockInTx(ctx context.Context, tx any, stock *WarehouseStock) error
	UpdateStockInTx(ctx context.Context, tx any, stock *WarehouseStock) error

	// 调拨操作
	SaveTransfer(ctx context.Context, transfer *StockTransfer) error
	FindTransferByNo(ctx context.Context, transferNo string) (*StockTransfer, error)
	UpdateTransfer(ctx context.Context, transfer *StockTransfer) error

	// 事务与 Barrier 管理
	BeginTx(ctx context.Context) any
	CommitTx(tx any) error
	RollbackTx(tx any) error
	WithTx(ctx context.Context, fn func(tx any) error) error
	WithBarrier(ctx context.Context, barrier any, fn func(tx any) error) error
}
