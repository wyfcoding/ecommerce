package application

import (
	"context"
	"ecommerce/internal/inventory/domain/entity"
	"ecommerce/internal/inventory/domain/repository"
	"errors"

	"log/slog"
)

type InventoryService struct {
	repo   repository.InventoryRepository
	logger *slog.Logger
}

func NewInventoryService(repo repository.InventoryRepository, logger *slog.Logger) *InventoryService {
	return &InventoryService{
		repo:   repo,
		logger: logger,
	}
}

// CreateInventory 创建库存
func (s *InventoryService) CreateInventory(ctx context.Context, skuID, productID uint64, totalStock, warningThreshold int32) (*entity.Inventory, error) {
	// Check if exists
	existing, err := s.repo.GetBySkuID(ctx, skuID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, errors.New("inventory already exists for this SKU")
	}

	inventory := entity.NewInventory(skuID, productID, totalStock, warningThreshold)
	if err := s.repo.Save(ctx, inventory); err != nil {
		s.logger.Error("failed to save inventory", "error", err)
		return nil, err
	}
	return inventory, nil
}

// GetInventory 获取库存
func (s *InventoryService) GetInventory(ctx context.Context, skuID uint64) (*entity.Inventory, error) {
	return s.repo.GetBySkuID(ctx, skuID)
}

// AddStock 增加库存
func (s *InventoryService) AddStock(ctx context.Context, skuID uint64, quantity int32, reason string) error {
	inventory, err := s.repo.GetBySkuID(ctx, skuID)
	if err != nil {
		return err
	}
	if inventory == nil {
		return errors.New("inventory not found")
	}

	if err := inventory.Add(quantity, reason); err != nil {
		return err
	}

	return s.repo.Save(ctx, inventory)
}

// DeductStock 扣减库存
func (s *InventoryService) DeductStock(ctx context.Context, skuID uint64, quantity int32, reason string) error {
	inventory, err := s.repo.GetBySkuID(ctx, skuID)
	if err != nil {
		return err
	}
	if inventory == nil {
		return errors.New("inventory not found")
	}

	if err := inventory.Deduct(quantity, reason); err != nil {
		return err
	}

	return s.repo.Save(ctx, inventory)
}

// LockStock 锁定库存
func (s *InventoryService) LockStock(ctx context.Context, skuID uint64, quantity int32, reason string) error {
	inventory, err := s.repo.GetBySkuID(ctx, skuID)
	if err != nil {
		return err
	}
	if inventory == nil {
		return errors.New("inventory not found")
	}

	if err := inventory.Lock(quantity, reason); err != nil {
		return err
	}

	return s.repo.Save(ctx, inventory)
}

// UnlockStock 解锁库存
func (s *InventoryService) UnlockStock(ctx context.Context, skuID uint64, quantity int32, reason string) error {
	inventory, err := s.repo.GetBySkuID(ctx, skuID)
	if err != nil {
		return err
	}
	if inventory == nil {
		return errors.New("inventory not found")
	}

	if err := inventory.Unlock(quantity, reason); err != nil {
		return err
	}

	return s.repo.Save(ctx, inventory)
}

// ConfirmDeduction 确认扣减
func (s *InventoryService) ConfirmDeduction(ctx context.Context, skuID uint64, quantity int32, reason string) error {
	inventory, err := s.repo.GetBySkuID(ctx, skuID)
	if err != nil {
		return err
	}
	if inventory == nil {
		return errors.New("inventory not found")
	}

	if err := inventory.ConfirmDeduction(quantity, reason); err != nil {
		return err
	}

	return s.repo.Save(ctx, inventory)
}

// ListInventories 获取库存列表
func (s *InventoryService) ListInventories(ctx context.Context, page, pageSize int) ([]*entity.Inventory, int64, error) {
	offset := (page - 1) * pageSize
	return s.repo.List(ctx, offset, pageSize)
}

// GetInventoryLogs 获取库存日志
func (s *InventoryService) GetInventoryLogs(ctx context.Context, inventoryID uint64, page, pageSize int) ([]*entity.InventoryLog, int64, error) {
	offset := (page - 1) * pageSize
	return s.repo.GetLogs(ctx, inventoryID, offset, pageSize)
}
