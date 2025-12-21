package application

import (
	"context"

	"github.com/wyfcoding/ecommerce/internal/inventory/domain"
	"github.com/wyfcoding/pkg/algorithm"
)

// InventoryService 是库存应用服务的门面。
type InventoryService struct {
	Manager *InventoryManager
	Query   *InventoryQuery
}

// NewInventoryService 创建库存服务实例。
func NewInventoryService(manager *InventoryManager, query *InventoryQuery) *InventoryService {
	return &InventoryService{
		Manager: manager,
		Query:   query,
	}
}

// CreateInventory 创建库存记录。
func (s *InventoryService) CreateInventory(ctx context.Context, skuID, productID, warehouseID uint64, totalStock, warningThreshold int32) (*domain.Inventory, error) {
	return s.Manager.CreateInventory(ctx, skuID, productID, warehouseID, totalStock, warningThreshold)
}

// GetInventory 获取库存详情。
func (s *InventoryService) GetInventory(ctx context.Context, skuID uint64) (*domain.Inventory, error) {
	return s.Query.GetInventory(ctx, skuID)
}

// AddStock 增加库存。
func (s *InventoryService) AddStock(ctx context.Context, skuID uint64, quantity int32, reason string) error {
	return s.Manager.AddStock(ctx, skuID, quantity, reason)
}

// DeductStock 扣减库存（直接扣减）。
func (s *InventoryService) DeductStock(ctx context.Context, skuID uint64, quantity int32, reason string) error {
	return s.Manager.DeductStock(ctx, skuID, quantity, reason)
}

// LockStock 锁定库存。
func (s *InventoryService) LockStock(ctx context.Context, skuID uint64, quantity int32, reason string) error {
	return s.Manager.LockStock(ctx, skuID, quantity, reason)
}

// UnlockStock 解锁库存。
func (s *InventoryService) UnlockStock(ctx context.Context, skuID uint64, quantity int32, reason string) error {
	return s.Manager.UnlockStock(ctx, skuID, quantity, reason)
}

// ConfirmDeduction 确认扣减（将锁定库存转为已扣减）。
func (s *InventoryService) ConfirmDeduction(ctx context.Context, skuID uint64, quantity int32, reason string) error {
	return s.Manager.ConfirmDeduction(ctx, skuID, quantity, reason)
}

// ListInventories 获取库存列表。
func (s *InventoryService) ListInventories(ctx context.Context, page, pageSize int) ([]*domain.Inventory, int64, error) {
	return s.Query.ListInventories(ctx, page, pageSize)
}

// GetInventoryLogs 获取库存变更日志。
func (s *InventoryService) GetInventoryLogs(ctx context.Context, inventoryID uint64, page, pageSize int) ([]*domain.InventoryLog, int64, error) {
	return s.Query.GetInventoryLogs(ctx, inventoryID, page, pageSize)
}

// AllocateStock 为订单分配库存。
func (s *InventoryService) AllocateStock(ctx context.Context, userLat, userLon float64, items []algorithm.OrderItem) ([]algorithm.AllocationResult, error) {
	return s.Manager.AllocateStock(ctx, userLat, userLon, items)
}
