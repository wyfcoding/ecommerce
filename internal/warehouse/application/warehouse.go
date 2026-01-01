package application

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/wyfcoding/ecommerce/internal/warehouse/domain"
)

// WarehouseService 结构体定义了仓库管理相关的应用服务 (外观模式)。
// 它协调 WarehouseManager 和 WarehouseQuery 处理仓库创建、库存管理、调拨等核心业务。
type WarehouseService struct {
	manager *WarehouseManager
	query   *WarehouseQuery
}

// NewWarehouseService 创建仓库服务门面实例。
func NewWarehouseService(manager *WarehouseManager, query *WarehouseQuery) *WarehouseService {
	return &WarehouseService{
		manager: manager,
		query:   query,
	}
}

// CreateWarehouse 创建一个新仓库记录。
func (s *WarehouseService) CreateWarehouse(ctx context.Context, code, name string) (*domain.Warehouse, error) {
	warehouse := &domain.Warehouse{
		Code:   code,
		Name:   name,
		Status: domain.WarehouseStatusInactive,
	}
	if err := s.manager.CreateWarehouse(ctx, warehouse); err != nil {
		return nil, err
	}
	return warehouse, nil
}

// GetWarehouse 根据ID获取仓库详情。
func (s *WarehouseService) GetWarehouse(ctx context.Context, id uint64) (*domain.Warehouse, error) {
	return s.query.GetWarehouseByID(ctx, id)
}

// UpdateStock 更新指定仓库和SKU的库存实物数量（全量或增量调整）。
func (s *WarehouseService) UpdateStock(ctx context.Context, warehouseID, skuID uint64, quantity int32) error {
	stock, err := s.query.GetStock(ctx, warehouseID, skuID)
	if err != nil {
		return err
	}
	if stock == nil {
		if quantity < 0 {
			return errors.New("insufficient stock: SKU not found in warehouse")
		}
		stock = &domain.WarehouseStock{
			WarehouseID: warehouseID,
			SkuID:       skuID,
			Stock:       0,
		}
	}

	if stock.Stock+quantity < 0 {
		return errors.New("insufficient stock: cannot reduce stock below zero")
	}

	stock.Stock += quantity
	return s.manager.AdjustStock(ctx, stock)
}

// CreateTransfer 创建一个库存调拨申请单。
func (s *WarehouseService) CreateTransfer(ctx context.Context, fromID, toID, skuID uint64, quantity int32, createdBy uint64) (*domain.StockTransfer, error) {
	stock, err := s.query.GetStock(ctx, fromID, skuID)
	if err != nil {
		return nil, err
	}
	if stock == nil || stock.AvailableStock() < quantity {
		return nil, errors.New("insufficient stock in source warehouse for transfer")
	}

	// 锁定库存。
	stock.LockedStock += quantity
	if err := s.manager.AdjustStock(ctx, stock); err != nil {
		return nil, err
	}

	transfer := &domain.StockTransfer{
		TransferNo:      fmt.Sprintf("T%d%d", fromID, time.Now().UnixNano()),
		FromWarehouseID: fromID,
		ToWarehouseID:   toID,
		SkuID:           skuID,
		Quantity:        quantity,
		Status:          domain.StockTransferStatusPending,
		CreatedBy:       createdBy,
	}

	if err := s.manager.CreateTransfer(ctx, transfer); err != nil {
		return nil, err
	}

	return transfer, nil
}

// CompleteTransfer 确认并完成一个库存调拨流程。
func (s *WarehouseService) CompleteTransfer(ctx context.Context, transferID uint64) error {
	transfer, err := s.query.GetTransferByID(ctx, transferID)
	if err != nil {
		return err
	}
	if transfer == nil {
		return errors.New("transfer not found")
	}
	if transfer.Status == domain.StockTransferStatusCompleted {
		return nil
	}

	// 1. 从源仓库实际扣除库存和解锁库存。
	sourceStock, err := s.query.GetStock(ctx, transfer.FromWarehouseID, transfer.SkuID)
	if err != nil {
		return err
	}
	if sourceStock == nil {
		return errors.New("source stock not found for transfer completion")
	}
	sourceStock.LockedStock -= transfer.Quantity
	sourceStock.Stock -= transfer.Quantity
	if err := s.manager.AdjustStock(ctx, sourceStock); err != nil {
		return err
	}

	// 2. 增加目标仓库的库存。
	destStock, err := s.query.GetStock(ctx, transfer.ToWarehouseID, transfer.SkuID)
	if err != nil {
		return err
	}
	if destStock == nil {
		destStock = &domain.WarehouseStock{
			WarehouseID: transfer.ToWarehouseID,
			SkuID:       transfer.SkuID,
			Stock:       0,
		}
	}
	destStock.Stock += transfer.Quantity
	if err := s.manager.AdjustStock(ctx, destStock); err != nil {
		return err
	}

	// 3. 更新调拨单状态。
	transfer.Status = domain.StockTransferStatusCompleted
	now := time.Now()
	transfer.CompletedAt = &now
	return s.manager.UpdateTransferStatus(ctx, transfer)
}

// GetTransfer 获取指定调拨单的详细信息。
func (s *WarehouseService) GetTransfer(ctx context.Context, id uint64) (*domain.StockTransfer, error) {
	return s.query.GetTransferByID(ctx, id)
}

// ListTransfers 获取调拨单列表（分页）。
func (s *WarehouseService) ListTransfers(ctx context.Context, fromID, toID uint64, status *domain.StockTransferStatus, page, pageSize int) ([]*domain.StockTransfer, int64, error) {
	offset := (page - 1) * pageSize
	return s.query.ListTransfers(ctx, fromID, toID, status, offset, pageSize)
}

// ListWarehouses 分页列出所有仓库。
func (s *WarehouseService) ListWarehouses(ctx context.Context, page, pageSize int) ([]*domain.Warehouse, int64, error) {
	offset := (page - 1) * pageSize
	return s.query.SearchWarehouses(ctx, nil, offset, pageSize)
}

// GetStock 获取指定SKU在指定仓库中的实时库存快照。
func (s *WarehouseService) GetStock(ctx context.Context, warehouseID, skuID uint64) (*domain.WarehouseStock, error) {
	return s.query.GetStock(ctx, warehouseID, skuID)
}

// DeductStock 核心交易接口：扣减库存 (支持 Saga 分布式事务的正向操作)。
func (s *WarehouseService) DeductStock(ctx context.Context, warehouseID, skuID uint64, quantity int32) error {
	stock, err := s.query.GetStock(ctx, warehouseID, skuID)
	if err != nil {
		return err
	}
	if stock == nil {
		return errors.New("stock not found for deduction")
	}
	if stock.Stock < quantity {
		return errors.New("insufficient stock for deduction")
	}

	stock.Stock -= quantity
	return s.manager.AdjustStock(ctx, stock)
}

// RevertStock 核心交易接口：回滚库存 (支持 Saga 分布式事务的逆向补偿操作)。
func (s *WarehouseService) RevertStock(ctx context.Context, warehouseID, skuID uint64, quantity int32) error {
	stock, err := s.query.GetStock(ctx, warehouseID, skuID)
	if err != nil {
		return err
	}
	if stock == nil {
		stock = &domain.WarehouseStock{
			WarehouseID: warehouseID,
			SkuID:       skuID,
			Stock:       0,
		}
	}

	stock.Stock += quantity
	return s.manager.AdjustStock(ctx, stock)
}

// GetOptimalWarehouse 返回地理位置最优（距离最近）且库存充足的发货仓库。
func (s *WarehouseService) GetOptimalWarehouse(ctx context.Context, skuID uint64, qty int32, lat, lon float64) (*domain.Warehouse, float64, int32, error) {
	return s.query.GetOptimalWarehouse(ctx, skuID, qty, lat, lon)
}
