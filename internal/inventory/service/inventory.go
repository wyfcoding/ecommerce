package service

import (
	"context"
	"encoding/json"
	"fmt"

	"go.uber.org/zap"

	"ecommerce/internal/inventory/model"
	"ecommerce/internal/inventory/repository"
)

// InventoryService 定义了库存服务的业务逻辑接口
type InventoryService interface {
	// gRPC 接口方法
	GetStockLevel(ctx context.Context, sku string, warehouseID uint) (*model.Inventory, error)
	AdjustStock(ctx context.Context, sku string, warehouseID uint, quantityChange int, movementType model.MovementType, reference, reason string) error

	// 消息队列消费者处理方法
	HandleOrderCreated(ctx context.Context, payload []byte) error
	HandleOrderCancelled(ctx context.Context, payload []byte) error
}

// inventoryService 是接口的具体实现
type inventoryService struct {
	repo   repository.InventoryRepository
	logger *zap.Logger
}

// NewInventoryService 创建一个新的 inventoryService 实例
func NewInventoryService(repo repository.InventoryRepository, logger *zap.Logger) InventoryService {
	return &inventoryService{repo: repo, logger: logger}
}

// GetStockLevel 获取库存水平
func (s *inventoryService) GetStockLevel(ctx context.Context, sku string, warehouseID uint) (*model.Inventory, error) {
	s.logger.Info("GetStockLevel called", zap.String("sku", sku), zap.Uint("warehouseID", warehouseID))
	inventory, err := s.repo.GetInventory(ctx, sku, warehouseID)
	if err != nil {
		s.logger.Error("Failed to get inventory", zap.Error(err))
		return nil, err
	}
	if inventory == nil {
		return nil, fmt.Errorf("库存记录不存在")
	}
	return inventory, nil
}

// AdjustStock 手动调整库存
func (s *inventoryService) AdjustStock(ctx context.Context, sku string, warehouseID uint, quantityChange int, movementType model.MovementType, reference, reason string) error {
	s.logger.Info("AdjustStock called", zap.String("sku", sku), zap.Int("quantityChange", quantityChange))
	return s.repo.AdjustStock(ctx, sku, warehouseID, quantityChange, movementType, reference, reason)
}

// OrderEventPayload 是从订单服务接收到的事件消息体结构
type OrderEventPayload struct {
	OrderSN string `json:"order_sn"`
	Items   []struct {
		SKU      string `json:"sku"`
		Quantity int    `json:"quantity"`
	} `json:"items"`
}

// HandleOrderCreated 处理订单创建事件，进行库存预留
func (s *inventoryService) HandleOrderCreated(ctx context.Context, payload []byte) error {
	var event OrderEventPayload
	if err := json.Unmarshal(payload, &event); err != nil {
		s.logger.Error("Failed to unmarshal order created event", zap.Error(err))
		return err
	}

	s.logger.Info("Handling order created event", zap.String("orderSN", event.OrderSN))

	// 伪代码: 假设所有商品都在同一个仓库
	const defaultWarehouseID = 1

	// TODO: 此处需要实现分布式事务（SAGA 模式）
	// 1. 遍历所有商品项，逐个预留库存
	for _, item := range event.Items {
		if err := s.repo.ReserveStock(ctx, item.SKU, defaultWarehouseID, item.Quantity); err != nil {
			s.logger.Error("Failed to reserve stock for item", zap.String("sku", item.SKU), zap.Error(err))
			// **补偿逻辑**: 如果某个商品预留失败，需要回滚已预留成功的商品库存
			// s.rollbackReservations(ctx, event.Items[:i])
			return fmt.Errorf("为 SKU %s 预留库存失败: %w", item.SKU, err)
		}
	}

	return nil
}

// HandleOrderCancelled 处理订单取消事件，回滚库存
func (s *inventoryService) HandleOrderCancelled(ctx context.Context, payload []byte) error {
	var event OrderEventPayload
	if err := json.Unmarshal(payload, &event); err != nil {
		return err
	}

	s.logger.Info("Handling order cancelled event", zap.String("orderSN", event.OrderSN))

	const defaultWarehouseID = 1

	// 1. 释放预留库存
	for _, item := range event.Items {
		if err := s.repo.ReleaseStock(ctx, item.SKU, defaultWarehouseID, item.Quantity); err != nil {
			s.logger.Error("Failed to release stock for item", zap.String("sku", item.SKU), zap.Error(err))
			// 补偿/告警逻辑
			return err
		}
	}

	// 2. 增加物理库存 (如果出库时是直接扣减物理库存)
	// 在我们的模型中，发货时才会扣减物理库存，所以取消时只需释放预留库存即可

	return nil
}