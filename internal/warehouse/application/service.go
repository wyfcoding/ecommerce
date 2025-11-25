package application

import (
	"context"
	"github.com/wyfcoding/ecommerce/internal/warehouse/domain/entity"
	"github.com/wyfcoding/ecommerce/internal/warehouse/domain/repository"
	"errors"
	"fmt"
	"time"

	"log/slog"
)

type WarehouseService struct {
	repo   repository.WarehouseRepository
	logger *slog.Logger
}

func NewWarehouseService(repo repository.WarehouseRepository, logger *slog.Logger) *WarehouseService {
	return &WarehouseService{
		repo:   repo,
		logger: logger,
	}
}

// CreateWarehouse 创建仓库
func (s *WarehouseService) CreateWarehouse(ctx context.Context, code, name string) (*entity.Warehouse, error) {
	warehouse := &entity.Warehouse{
		Code:   code,
		Name:   name,
		Status: entity.WarehouseStatusInactive,
	}
	if err := s.repo.SaveWarehouse(ctx, warehouse); err != nil {
		return nil, err
	}
	return warehouse, nil
}

// UpdateStock 更新库存 (Add or Deduct)
func (s *WarehouseService) UpdateStock(ctx context.Context, warehouseID, skuID uint64, quantity int32) error {
	stock, err := s.repo.GetStock(ctx, warehouseID, skuID)
	if err != nil {
		return err
	}
	if stock == nil {
		if quantity < 0 {
			return errors.New("insufficient stock")
		}
		stock = &entity.WarehouseStock{
			WarehouseID: warehouseID,
			SkuID:       skuID,
			Stock:       0,
		}
	}

	if stock.Stock+quantity < 0 {
		return errors.New("insufficient stock")
	}

	stock.Stock += quantity
	return s.repo.SaveStock(ctx, stock)
}

// CreateTransfer 创建调拨单
func (s *WarehouseService) CreateTransfer(ctx context.Context, fromID, toID, skuID uint64, quantity int32, createdBy uint64) (*entity.StockTransfer, error) {
	// Check stock in source warehouse
	stock, err := s.repo.GetStock(ctx, fromID, skuID)
	if err != nil {
		return nil, err
	}
	if stock == nil || stock.AvailableStock() < quantity {
		return nil, errors.New("insufficient stock in source warehouse")
	}

	// Lock stock
	stock.LockedStock += quantity
	if err := s.repo.SaveStock(ctx, stock); err != nil {
		return nil, err
	}

	transfer := &entity.StockTransfer{
		TransferNo:      fmt.Sprintf("T%d%d", fromID, time.Now().UnixNano()),
		FromWarehouseID: fromID,
		ToWarehouseID:   toID,
		SkuID:           skuID,
		Quantity:        quantity,
		Status:          entity.StockTransferStatusPending,
		CreatedBy:       createdBy,
	}

	if err := s.repo.SaveTransfer(ctx, transfer); err != nil {
		return nil, err
	}

	return transfer, nil
}

// CompleteTransfer 完成调拨
func (s *WarehouseService) CompleteTransfer(ctx context.Context, transferID uint64) error {
	transfer, err := s.repo.GetTransfer(ctx, transferID)
	if err != nil {
		return err
	}
	if transfer == nil {
		return errors.New("transfer not found")
	}
	if transfer.Status == entity.StockTransferStatusCompleted {
		return nil
	}

	// Deduct from source
	sourceStock, err := s.repo.GetStock(ctx, transfer.FromWarehouseID, transfer.SkuID)
	if err != nil {
		return err
	}
	sourceStock.LockedStock -= transfer.Quantity
	sourceStock.Stock -= transfer.Quantity
	if err := s.repo.SaveStock(ctx, sourceStock); err != nil {
		return err
	}

	// Add to destination
	destStock, err := s.repo.GetStock(ctx, transfer.ToWarehouseID, transfer.SkuID)
	if err != nil {
		return err
	}
	if destStock == nil {
		destStock = &entity.WarehouseStock{
			WarehouseID: transfer.ToWarehouseID,
			SkuID:       transfer.SkuID,
			Stock:       0,
		}
	}
	destStock.Stock += transfer.Quantity
	if err := s.repo.SaveStock(ctx, destStock); err != nil {
		return err
	}

	// Update transfer status
	transfer.Status = entity.StockTransferStatusCompleted
	now := time.Now()
	transfer.CompletedAt = &now
	return s.repo.SaveTransfer(ctx, transfer)
}

// ListWarehouses 获取仓库列表
func (s *WarehouseService) ListWarehouses(ctx context.Context, page, pageSize int) ([]*entity.Warehouse, int64, error) {
	offset := (page - 1) * pageSize
	return s.repo.ListWarehouses(ctx, nil, offset, pageSize)
}

// GetStock 获取库存
func (s *WarehouseService) GetStock(ctx context.Context, warehouseID, skuID uint64) (*entity.WarehouseStock, error) {
	return s.repo.GetStock(ctx, warehouseID, skuID)
}
