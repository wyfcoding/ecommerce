package application

import (
	"context"

	"github.com/wyfcoding/ecommerce/internal/orderoptimization/domain"
)

// OptimizationQueryService 处理订单优化的读操作。
type OptimizationQueryService struct {
	repo domain.OrderOptimizationRepository
}

// NewOptimizationQueryService 创建一个新的 OptimizationQueryService 实例。
func NewOptimizationQueryService(repo domain.OrderOptimizationRepository) *OptimizationQueryService {
	return &OptimizationQueryService{
		repo: repo,
	}
}

// GetMergedOrder 获取合并订单详情。
func (q *OptimizationQueryService) GetMergedOrder(ctx context.Context, id uint64) (*domain.MergedOrder, error) {
	return q.repo.GetMergedOrder(ctx, id)
}

// ListSplitOrders 获取拆分订单列表。
func (q *OptimizationQueryService) ListSplitOrders(ctx context.Context, originalOrderID uint64) ([]*domain.SplitOrder, error) {
	return q.repo.ListSplitOrders(ctx, originalOrderID)
}

// GetAllocationPlan 获取仓库分配计划详情。
func (q *OptimizationQueryService) GetAllocationPlan(ctx context.Context, orderID uint64) (*domain.WarehouseAllocationPlan, error) {
	return q.repo.GetAllocationPlan(ctx, orderID)
}
