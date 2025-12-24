package domain

import (
	"context"
)

// OrderOptimizationRepository 是订单优化模块的仓储接口。
type OrderOptimizationRepository interface {
	// MergedOrder
	SaveMergedOrder(ctx context.Context, order *MergedOrder) error
	GetMergedOrder(ctx context.Context, id uint64) (*MergedOrder, error)

	// SplitOrder
	SaveSplitOrder(ctx context.Context, order *SplitOrder) error
	ListSplitOrders(ctx context.Context, originalOrderID uint64) ([]*SplitOrder, error)

	// WarehouseAllocationPlan
	SaveAllocationPlan(ctx context.Context, plan *WarehouseAllocationPlan) error
	GetAllocationPlan(ctx context.Context, orderID uint64) (*WarehouseAllocationPlan, error)
}
