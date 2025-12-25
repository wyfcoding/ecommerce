package application

import (
	"context"

	"github.com/wyfcoding/ecommerce/internal/orderoptimization/domain"
)

// OrderOptimizationService 作为订单优化操作的门面。
type OrderOptimizationService struct {
	manager *OptimizationManager
	query   *OptimizationQuery
}

// NewOrderOptimizationService 创建订单优化服务门面实例。
func NewOrderOptimizationService(manager *OptimizationManager, query *OptimizationQuery) *OrderOptimizationService {
	return &OrderOptimizationService{
		manager: manager,
		query:   query,
	}
}

// --- 写操作（委托给 Manager）---

// MergeOrders 合并指定用户的多个待支付订单。
func (s *OrderOptimizationService) MergeOrders(ctx context.Context, userID uint64, orderIDs []uint64) (*domain.MergedOrder, error) {
	return s.manager.MergeOrders(ctx, userID, orderIDs)
}

// SplitOrder 将一个大订单拆分为多个子订单（按仓库或属性）。
func (s *OrderOptimizationService) SplitOrder(ctx context.Context, orderID uint64) ([]*domain.SplitOrder, error) {
	return s.manager.SplitOrder(ctx, orderID)
}

// AllocateWarehouse 为订单中的商品分配最合适的仓库。
func (s *OrderOptimizationService) AllocateWarehouse(ctx context.Context, orderID uint64) (*domain.WarehouseAllocationPlan, error) {
	return s.manager.AllocateWarehouse(ctx, orderID)
}

// --- 读操作（委托给 Query）---

// GetMergedOrder 获取合并订单的详细信息。
func (s *OrderOptimizationService) GetMergedOrder(ctx context.Context, id uint64) (*domain.MergedOrder, error) {
	return s.query.GetMergedOrder(ctx, id)
}

// ListSplitOrders 列出由指定原始订单拆分出来的所有子订单。
func (s *OrderOptimizationService) ListSplitOrders(ctx context.Context, originalOrderID uint64) ([]*domain.SplitOrder, error) {
	return s.query.ListSplitOrders(ctx, originalOrderID)
}

// GetAllocationPlan 获取订单的仓库分配计划。
func (s *OrderOptimizationService) GetAllocationPlan(ctx context.Context, orderID uint64) (*domain.WarehouseAllocationPlan, error) {
	return s.query.GetAllocationPlan(ctx, orderID)
}
