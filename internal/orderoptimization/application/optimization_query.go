package application

import (
	"context"

	"github.com/wyfcoding/ecommerce/internal/orderoptimization/domain"
)

// OptimizationQuery 处理订单优化的读操作。
type OptimizationQuery struct {
	repo domain.OrderOptimizationRepository
}

// NewOptimizationQuery creates a new OptimizationQuery instance.
func NewOptimizationQuery(repo domain.OrderOptimizationRepository) *OptimizationQuery {
	return &OptimizationQuery{
		repo: repo,
	}
}

// GetMergedOrder 获取合并订单详情。
func (q *OptimizationQuery) GetMergedOrder(ctx context.Context, id uint64) (*domain.MergedOrder, error) {
	return q.repo.GetMergedOrder(ctx, id)
}

// ListSplitOrders 获取拆分订单列表。
func (q *OptimizationQuery) ListSplitOrders(ctx context.Context, originalOrderID uint64) ([]*domain.SplitOrder, error) {
	return q.repo.ListSplitOrders(ctx, originalOrderID)
}

// GetAllocationPlan 获取仓库分配计划详情。
func (q *OptimizationQuery) GetAllocationPlan(ctx context.Context, orderID uint64) (*domain.WarehouseAllocationPlan, error) {
	return q.repo.GetAllocationPlan(ctx, orderID)
}
