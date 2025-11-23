package repository

import (
	"context"
	"ecommerce/internal/order_optimization/domain/entity"
)

// OrderOptimizationRepository 订单优化仓储接口
type OrderOptimizationRepository interface {
	// 合并订单
	SaveMergedOrder(ctx context.Context, order *entity.MergedOrder) error
	GetMergedOrder(ctx context.Context, id uint64) (*entity.MergedOrder, error)

	// 拆分订单
	SaveSplitOrder(ctx context.Context, order *entity.SplitOrder) error
	ListSplitOrders(ctx context.Context, originalOrderID uint64) ([]*entity.SplitOrder, error)

	// 仓库分配
	SaveAllocationPlan(ctx context.Context, plan *entity.WarehouseAllocationPlan) error
	GetAllocationPlan(ctx context.Context, orderID uint64) (*entity.WarehouseAllocationPlan, error)
}
