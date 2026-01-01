package application

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/wyfcoding/ecommerce/internal/inventory/domain"
	"github.com/wyfcoding/pkg/algorithm"
)

// InventoryManager 处理库存的写操作（增删改、锁定、分配）。
type InventoryManager struct {
	repo          domain.InventoryRepository
	warehouseRepo domain.WarehouseRepository
	allocator     *algorithm.WarehouseAllocator
	logger        *slog.Logger
	soldOutFilter *algorithm.CuckooFilter
	filterMu      sync.RWMutex
}

// NewInventoryManager 负责处理 NewInventory 相关的写操作和业务逻辑。
func NewInventoryManager(repo domain.InventoryRepository, warehouseRepo domain.WarehouseRepository, logger *slog.Logger) *InventoryManager {
	return &InventoryManager{
		repo:          repo,
		warehouseRepo: warehouseRepo,
		allocator:     algorithm.NewWarehouseAllocator(),
		logger:        logger,
		soldOutFilter: algorithm.NewCuckooFilter(100000),
	}
}

// IsSoldOutQuickCheck 本地快速检查是否售罄
func (m *InventoryManager) IsSoldOutQuickCheck(skuID uint64) bool {
	m.filterMu.RLock()
	defer m.filterMu.RUnlock()
	return m.soldOutFilter.Contains([]byte(fmt.Sprintf("%d", skuID)))
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

// executeWithRetry 执行带乐观锁重试的库存更新逻辑
func (m *InventoryManager) executeWithRetry(ctx context.Context, skuID uint64, fn func(*domain.Inventory) (*domain.InventoryLog, error)) error {
	maxRetries := 3
	for i := 0; i < maxRetries; i++ {
		inventory, err := m.repo.GetBySkuID(ctx, skuID)
		if err != nil {
			return err
		}
		if inventory == nil {
			return errors.New("inventory not found")
		}

		// 执行业务逻辑
		log, err := fn(inventory)
		if err != nil {
			return err
		}

		// 尝试保存（带版本检查）
		err = m.repo.SaveWithOptimisticLock(ctx, inventory)
		if err == nil {
			// 保存成功，记录日志
			if log != nil {
				if logErr := m.repo.SaveLog(ctx, log); logErr != nil {
					m.logger.WarnContext(ctx, "failed to save inventory log", "log", log, "error", logErr)
				}
			}
			return nil
		}

		// 如果不是乐观锁失败，直接返回错误
		if err.Error() != "optimistic lock failed" {
			return err
		}
		
		// 乐观锁失败，等待后重试
		time.Sleep(time.Millisecond * time.Duration(10*(i+1)))
	}
	return errors.New("concurrent update failed after retries")
}

// AddStock 增加库存。
func (m *InventoryManager) AddStock(ctx context.Context, skuID uint64, quantity int32, reason string) error {
	return m.executeWithRetry(ctx, skuID, func(inv *domain.Inventory) (*domain.InventoryLog, error) {
		log, err := inv.Add(quantity, reason)
		if err != nil {
			return nil, err
		}
		
		// 如果库存不再为0，从售罄过滤器中移除
		if inv.AvailableStock > 0 {
			m.filterMu.Lock()
			m.soldOutFilter.Delete([]byte(fmt.Sprintf("%d", skuID)))
			m.filterMu.Unlock()
		}
		return log, nil
	})
}

// DeductStock 扣减库存。
func (m *InventoryManager) DeductStock(ctx context.Context, skuID uint64, quantity int32, reason string) error {
	return m.executeWithRetry(ctx, skuID, func(inv *domain.Inventory) (*domain.InventoryLog, error) {
		log, err := inv.Deduct(quantity, reason)
		if err != nil {
			return nil, err
		}

		// 如果库存归零，加入售罄过滤器
		if inv.AvailableStock <= 0 {
			m.filterMu.Lock()
			m.soldOutFilter.Add([]byte(fmt.Sprintf("%d", skuID)))
			m.filterMu.Unlock()
		}
		return log, nil
	})
}

// LockStock 锁定库存。
func (m *InventoryManager) LockStock(ctx context.Context, skuID uint64, quantity int32, reason string) error {
	return m.executeWithRetry(ctx, skuID, func(inv *domain.Inventory) (*domain.InventoryLog, error) {
		return inv.Lock(quantity, reason)
	})
}

// UnlockStock 解锁库存。
func (m *InventoryManager) UnlockStock(ctx context.Context, skuID uint64, quantity int32, reason string) error {
	return m.executeWithRetry(ctx, skuID, func(inv *domain.Inventory) (*domain.InventoryLog, error) {
		return inv.Unlock(quantity, reason)
	})
}

// ConfirmDeduction 确认扣减。
func (m *InventoryManager) ConfirmDeduction(ctx context.Context, skuID uint64, quantity int32, reason string) error {
	return m.executeWithRetry(ctx, skuID, func(inv *domain.Inventory) (*domain.InventoryLog, error) {
		return inv.ConfirmDeduction(quantity, reason)
	})
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
