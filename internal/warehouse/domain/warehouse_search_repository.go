package domain

import (
	"context"
	"time"
)

// WarehouseSearchRepository 定义仓库搜索仓储接口（Elasticsearch）。
type WarehouseSearchRepository interface {
	// Warehouse
	IndexWarehouse(ctx context.Context, warehouse *Warehouse) error
	DeleteWarehouse(ctx context.Context, id uint64) error
	SearchWarehouses(ctx context.Context, code, name, province, city, status *string, offset, limit int, startTime, endTime *time.Time, sortBy string) ([]*Warehouse, int64, error)

	// Transfer
	IndexTransfer(ctx context.Context, transfer *StockTransfer) error
	DeleteTransfer(ctx context.Context, id uint64) error
	SearchTransfers(ctx context.Context, fromID, toID *uint64, status *string, offset, limit int, startTime, endTime *time.Time, sortBy string) ([]*StockTransfer, int64, error)
}
