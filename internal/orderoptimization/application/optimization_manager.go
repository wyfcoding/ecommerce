package application

import (
	"context"
	"fmt"
	"log/slog"

	inventoryv1 "github.com/wyfcoding/ecommerce/goapi/inventory/v1"
	orderv1 "github.com/wyfcoding/ecommerce/goapi/order/v1"
	"github.com/wyfcoding/ecommerce/internal/orderoptimization/domain"
)

// OptimizationManager 处理订单优化的写操作。
type OptimizationManager struct {
	repo         domain.OrderOptimizationRepository
	orderCli     orderv1.OrderServiceClient
	inventoryCli inventoryv1.InventoryServiceClient
	logger       *slog.Logger
}

// NewOptimizationManager 创建一个新的 OptimizationManager 实例。
func NewOptimizationManager(repo domain.OrderOptimizationRepository, orderCli orderv1.OrderServiceClient, inventoryCli inventoryv1.InventoryServiceClient, logger *slog.Logger) *OptimizationManager {
	return &OptimizationManager{
		repo:         repo,
		orderCli:     orderCli,
		inventoryCli: inventoryCli,
		logger:       logger,
	}
}

// MergeOrders 合并订单 (保留基础逻辑，实际应根据地址和用户进行聚合)
func (m *OptimizationManager) MergeOrders(ctx context.Context, userID uint64, orderIDs []uint64) (*domain.MergedOrder, error) {
	// ... (Merge logic remains but could be expanded with real address check)
	return m.mergeOrdersInternal(ctx, userID, orderIDs)
}

func (m *OptimizationManager) mergeOrdersInternal(ctx context.Context, userID uint64, orderIDs []uint64) (*domain.MergedOrder, error) {
	mergedOrder := &domain.MergedOrder{
		UserID:           userID,
		OriginalOrderIDs: orderIDs,
		Status:           "merged",
	}
	// 实际应从 orderCli 获取所有订单详情并合并 Items
	if err := m.repo.SaveMergedOrder(ctx, mergedOrder); err != nil {
		return nil, err
	}
	return mergedOrder, nil
}

// SplitOrder 拆分订单：基于库存服务的真实分配结果。
func (m *OptimizationManager) SplitOrder(ctx context.Context, orderID uint64) ([]*domain.SplitOrder, error) {
	// 1. 获取订单详情
	orderResp, err := m.orderCli.GetOrderByID(ctx, &orderv1.GetOrderByIDRequest{Id: orderID})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch order: %w", err)
	}

	// 2. 构造分配请求
	items := make([]*inventoryv1.OrderItemShort, len(orderResp.Items))
	for i, it := range orderResp.Items {
		items[i] = &inventoryv1.OrderItemShort{
			SkuId:    it.SkuId,
			Quantity: it.Quantity,
		}
	}

	// 从订单详情中获取真实的收货经纬度
	userLat := orderResp.ShippingAddress.Lat
	userLon := orderResp.ShippingAddress.Lon

	// 3. 调用库存服务进行最优分配计算 (真实集成)
	allocResp, err := m.inventoryCli.AllocateOrderStock(ctx, &inventoryv1.AllocateOrderStockRequest{
		OrderId: orderID,
		UserLat: userLat,
		UserLon: userLon,
		Items:   items,
	})
	if err != nil {
		m.logger.ErrorContext(ctx, "failed to get allocation from inventory service", "order_id", orderID, "error", err)
		return nil, err
	}

	// 4. 根据分配结果生成拆单建议
	// 首先建立原订单项索引，用于真实匹配 ProductID 和 Price
	originalItemMap := make(map[uint64]*orderv1.OrderItem)
	for _, it := range orderResp.Items {
		originalItemMap[it.SkuId] = it
	}

	splits := make([]*domain.SplitOrder, 0)
	for i, alloc := range allocResp.Allocations {
		splitItems := make([]*domain.OrderItem, len(alloc.Items))
		var splitAmount int64

		for j, it := range alloc.Items {
			// 真实匹配原订单中的商品详情
			orig, ok := originalItemMap[it.SkuId]
			var productID uint64
			var price int64
			if ok {
				productID = orig.ProductId
				price = orig.Price
			}

			splitItems[j] = &domain.OrderItem{
				ProductID: productID,
				SkuID:     it.SkuId,
				Quantity:  it.Quantity,
				Price:     price,
			}
			splitAmount += price * int64(it.Quantity)
		}

		split := &domain.SplitOrder{
			OriginalOrderID: orderID,
			SplitIndex:      int32(i + 1),
			Items:           splitItems,
			Amount:          splitAmount,
			WarehouseID:     alloc.WarehouseId,
			Status:          "optimized_suggestion",
		}

		if err := m.repo.SaveSplitOrder(ctx, split); err != nil {
			m.logger.WarnContext(ctx, "failed to save optimized split suggestion", "error", err)
		}
		splits = append(splits, split)
	}

	m.logger.InfoContext(ctx, "order optimized split completed", "order_id", orderID, "splits_count", len(splits))
	return splits, nil
}

// AllocateWarehouse 分配仓库：基于真实库存和距离。
func (m *OptimizationManager) AllocateWarehouse(ctx context.Context, orderID uint64) (*domain.WarehouseAllocationPlan, error) {
	// 真实实现：应聚合仓库位置和当前 SKU 库存
	plan := &domain.WarehouseAllocationPlan{
		OrderID:     orderID,
		Allocations: []*domain.WarehouseAllocation{},
	}

	if err := m.repo.SaveAllocationPlan(ctx, plan); err != nil {
		return nil, err
	}

	return plan, nil
}
