package application

import (
	"context"
	"log/slog"

	"github.com/wyfcoding/ecommerce/internal/inventory/domain"
)

// InventoryQuery 处理库存的读操作。
type InventoryQuery struct {
	repo          domain.InventoryRepository
	warehouseRepo domain.WarehouseRepository
	logger        *slog.Logger
}

// NewInventoryQuery 负责处理 NewInventory 相关的读操作和查询逻辑。
func NewInventoryQuery(repo domain.InventoryRepository, warehouseRepo domain.WarehouseRepository, logger *slog.Logger) *InventoryQuery {
	return &InventoryQuery{
		repo:          repo,
		warehouseRepo: warehouseRepo,
		logger:        logger,
	}
}

// GetInventory 获取指定SKU的库存记录。
func (q *InventoryQuery) GetInventory(ctx context.Context, skuID uint64) (*domain.Inventory, error) {
	return q.repo.GetBySkuID(ctx, skuID)
}

// ListInventories 获取库存列表。
func (q *InventoryQuery) ListInventories(ctx context.Context, page, pageSize int) ([]*domain.Inventory, int64, error) {
	offset := (page - 1) * pageSize
	return q.repo.List(ctx, offset, pageSize)
}

// GetInventoryLogs 获取指定库存的日志列表。
func (q *InventoryQuery) GetInventoryLogs(ctx context.Context, inventoryID uint64, page, pageSize int) ([]*domain.InventoryLog, int64, error) {
	offset := (page - 1) * pageSize
	return q.repo.GetLogs(ctx, inventoryID, offset, pageSize)
}
