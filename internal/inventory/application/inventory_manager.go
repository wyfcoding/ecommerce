package application

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/wyfcoding/ecommerce/internal/inventory/domain"
	orderv1 "github.com/wyfcoding/ecommerce/goapi/order/v1"
	"github.com/wyfcoding/pkg/algorithm"
)

// InventoryManager 处理库存的写操作（增删改、锁定、分配）。
type InventoryManager struct {
	repo           domain.InventoryRepository
	warehouseRepo  domain.WarehouseRepository
	allocator      *algorithm.WarehouseAllocator
	logger         *slog.Logger
	soldOutFilter  *algorithm.CuckooFilter
	filterMu       sync.RWMutex
	remoteOrderCli orderv1.OrderServiceClient
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

func (m *InventoryManager) SetRemoteOrderClient(cli orderv1.OrderServiceClient) {
	m.remoteOrderCli = cli
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

// DeleteInventory 删除库存记录。
func (m *InventoryManager) DeleteInventory(ctx context.Context, skuID uint64) error {
	if err := m.repo.Delete(ctx, skuID); err != nil {
		m.logger.ErrorContext(ctx, "failed to delete inventory", "sku_id", skuID, "error", err)
		return err
	}
	m.logger.InfoContext(ctx, "inventory deleted successfully", "sku_id", skuID)
	return nil
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

		// --- 架构增强：自动补货触发 (Cross-Project Interaction) ---
		if inv.AvailableStock < inv.WarningThreshold && m.remoteOrderCli != nil {
			m.logger.InfoContext(ctx, "low stock detected, triggering institutional replenishment", "sku_id", skuID, "stock", inv.AvailableStock)

			// 真实化逻辑：根据预警阈值动态计算补货量
			replenishQty := int32(inv.WarningThreshold * 2)
			if replenishQty == 0 {
				replenishQty = 100 // 默认保底
			}

			go func() {
				// 异步下单，不阻塞主流程
				replenishCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancel()

				// 调用电商自身的订单服务发起补货申请
				_, err := m.remoteOrderCli.CreateOrder(replenishCtx, &orderv1.CreateOrderRequest{
					UserId: 999999, // 系统补货账户 ID (uint64)
					Items: []*orderv1.OrderItemCreate{
						{
							ProductId: inv.ProductID, // 使用当前库存对象的 SPU ID
							SkuId:     inv.SkuID,
							Quantity:  replenishQty,
						},
					},
					Remark: fmt.Sprintf("Auto-replenishment for low stock SKU %d", inv.SkuID),
				})
				if err != nil {
					m.logger.Error("failed to place replenishment order", "sku_id", skuID, "error", err)
				} else {
					m.logger.Info("replenishment order placed successfully", "sku_id", skuID, "quantity", replenishQty)
				}
			}()
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

// HandleOrderTimeout 处理订单支付超时，自动释放库存。
func (m *InventoryManager) HandleOrderTimeout(ctx context.Context, event map[string]any) error {
	orderID := uint64(event["order_id"].(float64))
	userID := uint64(event["user_id"].(float64))
	items := event["items"].([]any)

	m.logger.InfoContext(ctx, "checking order timeout for stock release", "order_id", orderID)

	// 1. 调用 Order Service 检查当前状态
	if m.remoteOrderCli != nil {
		resp, err := m.remoteOrderCli.GetOrderByID(ctx, &orderv1.GetOrderByIDRequest{
			Id:     orderID,
			UserId: userID,
		})
		if err != nil {
			return fmt.Errorf("failed to check order status for ID %d: %w", orderID, err)
		}

		// 只有处于 PENDING_PAYMENT 或类似初始状态的订单才需要自动释放
		// 如果订单已经是 PAID, SHIPPED, COMPLETED 等，绝对不能释放库存
		if resp.Status != orderv1.OrderStatus_PENDING_PAYMENT {
			m.logger.InfoContext(ctx, "order already processed or paid, skipping stock release", "order_id", orderID, "status", resp.Status)
			return nil
		}
	}

	// 2. 逐项释放库存 (补偿 LockStock)
	for _, it := range items {
		itemMap := it.(map[string]any)
		skuID := uint64(itemMap["sku_id"].(float64))
		qty := int32(itemMap["quantity"].(float64))

		m.logger.InfoContext(ctx, "auto-unlocking stock for timeout", "order_id", orderID, "sku_id", skuID, "qty", qty)
		if err := m.UnlockStock(ctx, skuID, qty, fmt.Sprintf("Auto-release for timeout order %d", orderID)); err != nil {
			m.logger.ErrorContext(ctx, "failed to auto-unlock stock", "sku_id", skuID, "error", err)
		}
	}

	return nil
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
