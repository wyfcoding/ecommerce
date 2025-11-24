package application

import (
	"context"
	"ecommerce/internal/inventory/domain/entity"
	"ecommerce/internal/inventory/domain/repository"
	"ecommerce/pkg/algorithm"
	"errors"

	"log/slog"
)

type InventoryService struct {
	repo          repository.InventoryRepository
	warehouseRepo repository.WarehouseRepository
	allocator     *algorithm.WarehouseAllocator
	logger        *slog.Logger
}

func NewInventoryService(repo repository.InventoryRepository, warehouseRepo repository.WarehouseRepository, logger *slog.Logger) *InventoryService {
	return &InventoryService{
		repo:          repo,
		warehouseRepo: warehouseRepo,
		allocator:     algorithm.NewWarehouseAllocator(),
		logger:        logger,
	}
}

// CreateInventory 创建库存
func (s *InventoryService) CreateInventory(ctx context.Context, skuID, productID, warehouseID uint64, totalStock, warningThreshold int32) (*entity.Inventory, error) {
	// Check if exists
	existing, err := s.repo.GetBySkuID(ctx, skuID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, errors.New("inventory already exists for this SKU")
	}

	inventory := entity.NewInventory(skuID, productID, warehouseID, totalStock, warningThreshold)
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

// AllocateStock 分配库存（使用仓库分配算法）
func (s *InventoryService) AllocateStock(ctx context.Context, userLat, userLon float64, items []algorithm.OrderItem) ([]algorithm.AllocationResult, error) {
	// 1. Fetch all warehouses
	warehouses, err := s.warehouseRepo.ListAll(ctx)
	if err != nil {
		return nil, err
	}

	// 2. Fetch inventory for items
	skuIDs := make([]uint64, len(items))
	for i, item := range items {
		skuIDs[i] = item.SkuID
	}
	inventories, err := s.repo.GetBySkuIDs(ctx, skuIDs)
	if err != nil {
		return nil, err
	}

	// 3. Prepare data for allocator
	// map[warehouseID]map[skuID]*WarehouseInfo
	warehouseMap := make(map[uint64]map[uint64]*algorithm.WarehouseInfo)

	// Helper to find warehouse by ID
	findWarehouse := func(id uint64) *entity.Warehouse {
		for _, w := range warehouses {
			if w.ID == uint(id) { // Assuming ID types match, but gorm.Model ID is uint
				return w
			}
		}
		return nil
	}

	for _, inv := range inventories {
		w := findWarehouse(inv.WarehouseID)
		if w == nil {
			continue // Should not happen if data consistency is maintained
		}

		if _, exists := warehouseMap[inv.WarehouseID]; !exists {
			warehouseMap[inv.WarehouseID] = make(map[uint64]*algorithm.WarehouseInfo)
		}

		warehouseMap[inv.WarehouseID][inv.SkuID] = &algorithm.WarehouseInfo{
			ID:       uint64(w.ID),
			Lat:      w.Lat,
			Lon:      w.Lon,
			Stock:    inv.AvailableStock,
			Priority: w.Priority,
			ShipCost: w.ShipCost,
		}
	}

	// 4. Run allocation algorithm
	results := s.allocator.AllocateOptimal(userLat, userLon, items, warehouseMap)

	// 5. Reserve stock based on results (Simplified: just return results, actual reservation should happen in transaction)
	// For this task, we just return the allocation plan.
	// In a real system, we would lock the stock here.

	return results, nil
}
