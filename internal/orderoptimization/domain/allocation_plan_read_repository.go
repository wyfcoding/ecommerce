package domain

import "context"

// AllocationPlanReadRepository 定义仓库分配计划读模型仓储接口（Redis）。
type AllocationPlanReadRepository interface {
	Save(ctx context.Context, plan *WarehouseAllocationPlan) error
	GetByOrderID(ctx context.Context, orderID uint64) (*WarehouseAllocationPlan, error)
	DeleteByOrderID(ctx context.Context, orderID uint64) error
}
