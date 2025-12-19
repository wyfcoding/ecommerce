package application

import (
	"context"

	"github.com/wyfcoding/ecommerce/internal/warehouse/domain"
)

// WarehouseQuery 处理仓库模块的查询操作。
type WarehouseQuery struct {
	repo domain.WarehouseRepository
}

// NewWarehouseQuery 创建并返回一个新的 WarehouseQuery 实例。
func NewWarehouseQuery(repo domain.WarehouseRepository) *WarehouseQuery {
	return &WarehouseQuery{repo: repo}
}

// GetWarehouseByID 根据ID获取仓库详情。
func (q *WarehouseQuery) GetWarehouseByID(ctx context.Context, id uint64) (*domain.Warehouse, error) {
	return q.repo.GetWarehouse(ctx, id)
}

// GetWarehouseByCode 根据代码获取仓库详情。
func (q *WarehouseQuery) GetWarehouseByCode(ctx context.Context, code string) (*domain.Warehouse, error) {
	return q.repo.GetWarehouseByCode(ctx, code)
}

// SearchWarehouses 搜索仓库。
func (q *WarehouseQuery) SearchWarehouses(ctx context.Context, status *domain.WarehouseStatus, offset, limit int) ([]*domain.Warehouse, int64, error) {
	return q.repo.ListWarehouses(ctx, status, offset, limit)
}

// GetStock 获取库存信息。
func (q *WarehouseQuery) GetStock(ctx context.Context, warehouseID, skuID uint64) (*domain.WarehouseStock, error) {
	return q.repo.GetStock(ctx, warehouseID, skuID)
}

// ListWarehouseStocks 列出仓库的所有库存。
func (q *WarehouseQuery) ListWarehouseStocks(ctx context.Context, warehouseID uint64, offset, limit int) ([]*domain.WarehouseStock, int64, error) {
	return q.repo.ListStocks(ctx, warehouseID, offset, limit)
}

// GetTransferByID 获取调拨单详情。
func (q *WarehouseQuery) GetTransferByID(ctx context.Context, id uint64) (*domain.StockTransfer, error) {
	return q.repo.GetTransfer(ctx, id)
}

// ListTransfers 列出调拨单。
func (q *WarehouseQuery) ListTransfers(ctx context.Context, fromWH, toWH uint64, status *domain.StockTransferStatus, offset, limit int) ([]*domain.StockTransfer, int64, error) {
	return q.repo.ListTransfers(ctx, fromWH, toWH, status, offset, limit)
}
