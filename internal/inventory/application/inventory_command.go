package application

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	orderv1 "github.com/wyfcoding/ecommerce/goapi/order/v1"
	"github.com/wyfcoding/ecommerce/internal/inventory/domain"
	algorithm "github.com/wyfcoding/pkg/algorithm/optimization"
	"github.com/wyfcoding/pkg/algorithm/structures"
	"github.com/wyfcoding/pkg/eventsourcing"
)

// InventoryCommandService 处理库存的写操作，集成了乐观锁重试、布隆过滤器预检及领域事件发布。
type InventoryCommandService struct {
	repo           domain.InventoryRepository                    // 库存仓储
	warehouseRepo  domain.WarehouseRepository                    // 仓库仓储
	publisher      domain.EventPublisher                         // 事件发布者
	eventStore     domain.EventStore                             // 事件存储
	allocator      *algorithm.WarehouseAllocator                 // 最优库存分配算法引擎
	logger         *slog.Logger                                  // 日志记录器
	soldOutFilter  *structures.CuckooFilter[structures.ByteHash] // 布隆/布谷鸟过滤器，用于高并发下的售罄快速判定
	filterMu       sync.RWMutex                                  // 保护过滤器的并发安全
	remoteOrderCli orderv1.OrderServiceClient                    // 远程订单服务客户端，用于触发自动补货
}

// NewInventoryCommandService 构造函数。
func NewInventoryCommandService(
	repo domain.InventoryRepository,
	warehouseRepo domain.WarehouseRepository,
	publisher domain.EventPublisher,
	eventStore domain.EventStore,
	logger *slog.Logger,
) (*InventoryCommandService, error) {
	filter, err := structures.NewCuckooFilter[structures.ByteHash](100000)
	if err != nil {
		return nil, err
	}

	return &InventoryCommandService{
		repo:          repo,
		warehouseRepo: warehouseRepo,
		publisher:     publisher,
		eventStore:    eventStore,
		allocator:     algorithm.NewWarehouseAllocator(),
		logger:        logger,
		soldOutFilter: filter,
	}, nil
}

func (m *InventoryCommandService) SetRemoteOrderClient(cli orderv1.OrderServiceClient) {
	m.remoteOrderCli = cli
}

// CreateInventory 创建一个新的库存记录。
func (m *InventoryCommandService) CreateInventory(ctx context.Context, skuID, productID, warehouseID uint64, totalStock, warningThreshold int32) (*domain.Inventory, error) {
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
func (m *InventoryCommandService) DeleteInventory(ctx context.Context, skuID uint64) error {
	if err := m.repo.Delete(ctx, skuID); err != nil {
		m.logger.ErrorContext(ctx, "failed to delete inventory", "sku_id", skuID, "error", err)
		return err
	}
	m.logger.InfoContext(ctx, "inventory deleted successfully", "sku_id", skuID)
	return nil
}

// executeWithRetry 执行带乐观锁重试的库存更新逻辑
func (m *InventoryCommandService) executeWithRetry(ctx context.Context, skuID uint64, fn func(*domain.Inventory) (*domain.InventoryLog, any, error)) error {
	maxRetries := 3
	for i := range maxRetries {
		inventory, err := m.repo.GetBySkuID(ctx, skuID)
		if err != nil {
			return err
		}
		if inventory == nil {
			return errors.New("inventory not found")
		}

		// 执行业务逻辑
		log, event, err := fn(inventory)
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
			// 持久化事件
			if m.eventStore != nil {
				events := inventory.GetUncommittedEvents()
				if len(events) > 0 {
					if saveErr := m.eventStore.Save(ctx, events); saveErr != nil {
						m.logger.WarnContext(ctx, "failed to save inventory events", "sku_id", skuID, "error", saveErr)
					}
					inventory.MarkCommitted()
				}
			}
			// 发布事件
			if event != nil {
				topic := m.getTopicForEvent(event)
				if topic != "" {
					_ = m.publisher.Publish(ctx, topic, fmt.Sprintf("%d", skuID), event)
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

func (m *InventoryCommandService) getTopicForEvent(event any) string {
	switch event.(type) {
	case *domain.StockLockedEvent:
		return "inventory.stock.locked"
	case *domain.StockUnlockedEvent:
		return "inventory.stock.unlocked"
	case *domain.StockDeductedEvent:
		return "inventory.stock.deducted"
	case *domain.StockAddedEvent:
		return "inventory.stock.added"
	case *domain.StockWarningEvent:
		return "inventory.stock.warning"
	}
	return ""
}

// AddStock 增加库存。
func (m *InventoryCommandService) AddStock(ctx context.Context, skuID uint64, quantity int32, reason string) error {
	return m.executeWithRetry(ctx, skuID, func(inv *domain.Inventory) (*domain.InventoryLog, any, error) {
		log, err := inv.Add(quantity, reason)
		if err != nil {
			return nil, nil, err
		}

		// 如果库存不再为0，从售罄过滤器中移除
		if inv.AvailableStock > 0 {
			m.filterMu.Lock()
			m.soldOutFilter.Delete(structures.ByteHash(fmt.Appendf(nil, "%d", skuID)))
			m.filterMu.Unlock()
		}

		// 事件已经在领域模型中通过 ApplyChange/Apply 生成并处理
		// 我们从聚合根中获取未提交的事件进行发布
		events := inv.GetUncommittedEvents()
		var lastEvent any
		if len(events) > 0 {
			lastEvent = events[len(events)-1]
		}

		return log, lastEvent, nil
	})
}

// DeductStock 扣减库存。
func (m *InventoryCommandService) DeductStock(ctx context.Context, skuID uint64, quantity int32, reason string) error {
	return m.executeWithRetry(ctx, skuID, func(inv *domain.Inventory) (*domain.InventoryLog, any, error) {
		log, err := inv.Deduct(quantity, reason)
		if err != nil {
			return nil, nil, err
		}

		// 如果库存归零，加入售罄过滤器
		if inv.AvailableStock <= 0 {
			m.filterMu.Lock()
			m.soldOutFilter.Add(structures.ByteHash(fmt.Appendf(nil, "%d", skuID)))
			m.filterMu.Unlock()
		}

		// 从聚合根中获取事件
		events := inv.GetUncommittedEvents()
		var lastEvent any
		if len(events) > 0 {
			lastEvent = events[len(events)-1]
		}

		// 检查预警 (这也可以由领域层抛出事件，但在此保留应用层检查逻辑)
		if inv.AvailableStock < inv.WarningThreshold {
			base := eventsourcing.NewBaseEvent(domain.StockWarningEventType, inv.GetID(), inv.AggregateRoot.Version())
			warningEvent := &domain.StockWarningEvent{
				BaseEvent:      base,
				SkuID:          skuID,
				AvailableStock: inv.AvailableStock,
				Threshold:      inv.WarningThreshold,
				Timestamp:      base.Timestamp,
			}
			_ = m.publisher.Publish(ctx, "inventory.stock.warning", fmt.Sprintf("%d", skuID), warningEvent)
		}

		return log, lastEvent, nil
	})
}

// LockStock 锁定库存。
func (m *InventoryCommandService) LockStock(ctx context.Context, skuID uint64, quantity int32, reason string) error {
	return m.executeWithRetry(ctx, skuID, func(inv *domain.Inventory) (*domain.InventoryLog, any, error) {
		log, err := inv.Lock(quantity, reason)
		if err != nil {
			return nil, nil, err
		}

		events := inv.GetUncommittedEvents()
		var lastEvent any
		if len(events) > 0 {
			lastEvent = events[len(events)-1]
		}
		return log, lastEvent, nil
	})
}

// UnlockStock 解锁库存。
func (m *InventoryCommandService) UnlockStock(ctx context.Context, skuID uint64, quantity int32, reason string) error {
	return m.executeWithRetry(ctx, skuID, func(inv *domain.Inventory) (*domain.InventoryLog, any, error) {
		log, err := inv.Unlock(quantity, reason)
		if err != nil {
			return nil, nil, err
		}

		events := inv.GetUncommittedEvents()
		var lastEvent any
		if len(events) > 0 {
			lastEvent = events[len(events)-1]
		}
		return log, lastEvent, nil
	})
}

// HandleOrderTimeout 处理订单支付超时，自动释放库存。
func (m *InventoryCommandService) HandleOrderTimeout(ctx context.Context, event map[string]any) error {
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

		if resp.Status == orderv1.OrderStatus_PAID ||
			resp.Status == orderv1.OrderStatus_SHIPPED ||
			resp.Status == orderv1.OrderStatus_DELIVERED ||
			resp.Status == orderv1.OrderStatus_COMPLETED {
			m.logger.InfoContext(ctx, "order already paid or processed, skipping stock release", "order_id", orderID, "status", resp.Status)
			return nil
		}

		m.logger.WarnContext(ctx, "order unpaid or stuck, proceeding with stock release", "order_id", orderID, "status", resp.Status)
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
func (m *InventoryCommandService) ConfirmDeduction(ctx context.Context, skuID uint64, quantity int32, reason string) error {
	return m.executeWithRetry(ctx, skuID, func(inv *domain.Inventory) (*domain.InventoryLog, any, error) {
		log, err := inv.ConfirmDeduction(quantity, reason)
		if err != nil {
			return nil, nil, err
		}

		events := inv.GetUncommittedEvents()
		var lastEvent any
		if len(events) > 0 {
			lastEvent = events[len(events)-1]
		}
		return log, lastEvent, nil
	})
}

// AllocateStock 分配库存。
func (m *InventoryCommandService) AllocateStock(ctx context.Context, userLat, userLon float64, items []algorithm.OrderItem) ([]algorithm.AllocationResult, error) {
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

	result := m.allocator.AllocateOptimal(userLat, userLon, items, warehouseMap)
	m.logger.InfoContext(ctx, "stock allocation optimization completed", "items_count", len(items), "allocations", len(result))
	return result, nil
}
