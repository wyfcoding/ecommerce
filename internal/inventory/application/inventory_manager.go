package application

import (
	"context"
	"errors"
	"log/slog"

	"github.com/wyfcoding/ecommerce/internal/inventory/domain"
	"github.com/wyfcoding/pkg/algorithm"
)

// InventoryManager 处理库存的写操作（增删改、锁定、分配）。
type InventoryManager struct {
	repo          domain.InventoryRepository
	warehouseRepo domain.WarehouseRepository
	allocator     *algorithm.WarehouseAllocator
	logger        *slog.Logger
}

// NewInventoryManager 负责处理 NewInventory 相关的写操作和业务逻辑。
func NewInventoryManager(repo domain.InventoryRepository, warehouseRepo domain.WarehouseRepository, logger *slog.Logger) *InventoryManager {
	return &InventoryManager{
		repo:          repo,
		warehouseRepo: warehouseRepo,
		allocator:     algorithm.NewWarehouseAllocator(),
		logger:        logger,
	}
}

// CreateInventory 创建一个新的库存记录。
func (m *InventoryManager) CreateInventory(ctx context.Context, skuID, productID, warehouseID uint64, totalStock, warningThreshold int32) (*domain.Inventory, error) {
	existing, err := m.repo.GetBySkuID(ctx, skuID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, errors.New("inventory already exists for this SKU")
	}

	inventory := domain.NewInventory(skuID, productID, warehouseID, totalStock, warningThreshold)
	if err := m.repo.Save(ctx, inventory); err != nil {
		m.logger.ErrorContext(ctx, "failed to save inventory", "sku_id", skuID, "error", err)
		return nil, err
	}
	m.logger.InfoContext(ctx, "inventory created successfully", "inventory_id", inventory.ID, "sku_id", skuID)
	return inventory, nil
}

// AddStock 增加库存。
func (m *InventoryManager) AddStock(ctx context.Context, skuID uint64, quantity int32, reason string) error {
	inventory, err := m.repo.GetBySkuID(ctx, skuID)
	if err != nil {
		return err
	}
	if inventory == nil {
		return errors.New("inventory not found")
	}

	if err := inventory.Add(quantity, reason); err != nil {
		return err
	}

	if err := m.repo.Save(ctx, inventory); err != nil {
		m.logger.ErrorContext(ctx, "failed to add stock", "sku_id", skuID, "error", err)
		return err
	}
	return nil
}

// DeductStock 扣减库存。
func (m *InventoryManager) DeductStock(ctx context.Context, skuID uint64, quantity int32, reason string) error {
	inventory, err := m.repo.GetBySkuID(ctx, skuID)
	if err != nil {
		return err
	}
	if inventory == nil {
		return errors.New("inventory not found")
	}

	if err := inventory.Deduct(quantity, reason); err != nil {
		return err
	}

	if err := m.repo.Save(ctx, inventory); err != nil {
		m.logger.ErrorContext(ctx, "failed to deduct stock", "sku_id", skuID, "error", err)
		return err
	}
	return nil
}

// LockStock 锁定库存。
func (m *InventoryManager) LockStock(ctx context.Context, skuID uint64, quantity int32, reason string) error {
	inventory, err := m.repo.GetBySkuID(ctx, skuID)
	if err != nil {
		return err
	}
	if inventory == nil {
		return errors.New("inventory not found")
	}

	if err := inventory.Lock(quantity, reason); err != nil {
		return err
	}

	if err := m.repo.Save(ctx, inventory); err != nil {
		m.logger.ErrorContext(ctx, "failed to lock stock", "sku_id", skuID, "error", err)
		return err
	}
	return nil
}

// UnlockStock 解锁库存。
func (m *InventoryManager) UnlockStock(ctx context.Context, skuID uint64, quantity int32, reason string) error {
	inventory, err := m.repo.GetBySkuID(ctx, skuID)
	if err != nil {
		return err
	}
	if inventory == nil {
		return errors.New("inventory not found")
	}

	if err := inventory.Unlock(quantity, reason); err != nil {
		return err
	}

	if err := m.repo.Save(ctx, inventory); err != nil {
		m.logger.ErrorContext(ctx, "failed to unlock stock", "sku_id", skuID, "error", err)
		return err
	}
	return nil
}

// ConfirmDeduction 确认扣减。
func (m *InventoryManager) ConfirmDeduction(ctx context.Context, skuID uint64, quantity int32, reason string) error {
	inventory, err := m.repo.GetBySkuID(ctx, skuID)
	if err != nil {
		return err
	}
	if inventory == nil {
		return errors.New("inventory not found")
	}

	if err := inventory.ConfirmDeduction(quantity, reason); err != nil {
		return err
	}

	if err := m.repo.Save(ctx, inventory); err != nil {
		m.logger.ErrorContext(ctx, "failed to confirm deduction", "sku_id", skuID, "error", err)
		return err
	}
	return nil
}

// AllocateStock 分配库存。
func (m *InventoryManager) AllocateStock(ctx context.Context, userLat, userLon float64, items []algorithm.OrderItem) ([]algorithm.AllocationResult, error) {
	warehouses, err := m.warehouseRepo.ListAll(ctx)
	if err != nil {
		return nil, err
	}

	skuIDs := make([]uint64, len(items))
	for i, item := range items {
		skuIDs[i] = item.SkuID
	}
	inventories, err := m.repo.GetBySkuIDs(ctx, skuIDs)
	if err != nil {
		return nil, err
	}

	warehouseMap := make(map[uint64]map[uint64]*algorithm.WarehouseInfo)
	findWarehouse := func(id uint64) *domain.Warehouse {
		for _, w := range warehouses {
			if w.ID == uint(id) {
				return w
			}
		}
		return nil
	}

	for _, inv := range inventories {
		w := findWarehouse(inv.WarehouseID)
		if w == nil {
			continue
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

	return m.allocator.AllocateOptimal(userLat, userLon, items, warehouseMap), nil
}
