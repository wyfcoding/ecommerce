package application

import (
	"context"

	"github.com/wyfcoding/ecommerce/internal/warehouse/domain"
)

// WarehouseService 是仓库应用服务的门面。
type WarehouseService struct {
	Command *WarehouseCommandService
	Query   *WarehouseQueryService
}

// NewWarehouseService 构造函数。
func NewWarehouseService(command *WarehouseCommandService, query *WarehouseQueryService) *WarehouseService {
	return &WarehouseService{
		Command: command,
		Query:   query,
	}
}

func (s *WarehouseService) CreateWarehouse(ctx context.Context, code, name string) (*domain.Warehouse, error) {
	return s.Command.CreateWarehouse(ctx, &domain.Warehouse{Code: code, Name: name})
}

func (s *WarehouseService) ListWarehouses(ctx context.Context, page, pageSize int) ([]*domain.Warehouse, int64, error) {
	return s.Query.ListWarehouses(ctx, page, pageSize)
}

func (s *WarehouseService) GetWarehouse(ctx context.Context, id uint64) (*domain.Warehouse, error) {
	return s.Query.GetWarehouse(ctx, id)
}

func (s *WarehouseService) UpdateStock(ctx context.Context, warehouseID, skuID uint64, qty int32) error {
	return s.Command.AdjustStock(ctx, warehouseID, skuID, qty, "Service Update")
}

func (s *WarehouseService) DeductStock(ctx context.Context, barrier any, orderID, skuID, warehouseID uint64, qty int32) error {
	return s.Command.DeductStock(ctx, barrier, orderID, skuID, warehouseID, qty)
}

func (s *WarehouseService) RevertStock(ctx context.Context, barrier any, orderID, skuID, warehouseID uint64, qty int32) error {
	return s.Command.RevertStock(ctx, barrier, orderID, skuID, warehouseID, qty)
}

func (s *WarehouseService) GetOptimalWarehouse(ctx context.Context, skuID uint64, qty int32, lat, lon float64) (*domain.Warehouse, float64, int32, error) {
	return s.Query.GetOptimalWarehouse(ctx, skuID, qty, lat, lon)
}

func (s *WarehouseService) GetTransfer(ctx context.Context, id uint64) (*domain.StockTransfer, error) {
	return s.Query.GetTransfer(ctx, id)
}

func (s *WarehouseService) ListTransfers(ctx context.Context, fromID, toID uint64, status *string, page, pageSize int) ([]*domain.StockTransfer, int64, error) {
	return s.Query.ListTransfers(ctx, fromID, toID, status, page, pageSize)
}

func (s *WarehouseService) CreateTransfer(ctx context.Context, fromID, toID, skuID uint64, qty int32, createdBy uint64) (*domain.StockTransfer, error) {
	return s.Command.CreateTransfer(ctx, fromID, toID, skuID, qty, createdBy)
}

func (s *WarehouseService) CompleteTransfer(ctx context.Context, id uint64) error {
	return s.Command.CompleteTransfer(ctx, id)
}
