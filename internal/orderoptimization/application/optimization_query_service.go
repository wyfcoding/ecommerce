package application

import (
	"context"
	"log/slog"

	"github.com/wyfcoding/ecommerce/internal/orderoptimization/domain"
)

// OptimizationQueryService 处理订单优化的读操作。
type OptimizationQueryService struct {
	repo               domain.OrderOptimizationRepository
	mergedReadRepo     domain.MergedOrderReadRepository
	splitReadRepo      domain.SplitOrderReadRepository
	allocationReadRepo domain.AllocationPlanReadRepository
	splitSearchRepo    domain.SplitOrderSearchRepository
	logger             *slog.Logger
}

// NewOptimizationQueryService 创建一个新的 OptimizationQueryService 实例。
func NewOptimizationQueryService(
	repo domain.OrderOptimizationRepository,
	mergedReadRepo domain.MergedOrderReadRepository,
	splitReadRepo domain.SplitOrderReadRepository,
	allocationReadRepo domain.AllocationPlanReadRepository,
	splitSearchRepo domain.SplitOrderSearchRepository,
	logger *slog.Logger,
) *OptimizationQueryService {
	return &OptimizationQueryService{
		repo:               repo,
		mergedReadRepo:     mergedReadRepo,
		splitReadRepo:      splitReadRepo,
		allocationReadRepo: allocationReadRepo,
		splitSearchRepo:    splitSearchRepo,
		logger:             logger,
	}
}

// GetMergedOrder 获取合并订单详情。
func (q *OptimizationQueryService) GetMergedOrder(ctx context.Context, id uint64) (*domain.MergedOrder, error) {
	if q.mergedReadRepo != nil {
		if cached, err := q.mergedReadRepo.GetByID(ctx, id); err == nil && cached != nil {
			return cached, nil
		}
	}
	order, err := q.repo.GetMergedOrder(ctx, id)
	if err != nil {
		return nil, err
	}
	if order != nil && q.mergedReadRepo != nil {
		_ = q.mergedReadRepo.Save(ctx, order)
	}
	return order, nil
}

// ListSplitOrders 获取拆分订单列表。
func (q *OptimizationQueryService) ListSplitOrders(ctx context.Context, originalOrderID uint64) ([]*domain.SplitOrder, error) {
	if q.splitSearchRepo != nil {
		list, _, err := q.splitSearchRepo.SearchByOriginalOrderID(ctx, originalOrderID, 0, 200)
		if err == nil {
			return list, nil
		}
		if q.logger != nil {
			q.logger.WarnContext(ctx, "split order search fallback to cache/mysql", "error", err)
		}
	}
	if q.splitReadRepo != nil {
		if cached, err := q.splitReadRepo.GetByOriginalOrderID(ctx, originalOrderID); err == nil && cached != nil {
			return cached, nil
		}
	}
	list, err := q.repo.ListSplitOrders(ctx, originalOrderID)
	if err != nil {
		return nil, err
	}
	if q.splitReadRepo != nil {
		_ = q.splitReadRepo.Save(ctx, originalOrderID, list)
	}
	return list, nil
}

// GetAllocationPlan 获取仓库分配计划详情。
func (q *OptimizationQueryService) GetAllocationPlan(ctx context.Context, orderID uint64) (*domain.WarehouseAllocationPlan, error) {
	if q.allocationReadRepo != nil {
		if cached, err := q.allocationReadRepo.GetByOrderID(ctx, orderID); err == nil && cached != nil {
			return cached, nil
		}
	}
	plan, err := q.repo.GetAllocationPlan(ctx, orderID)
	if err != nil {
		return nil, err
	}
	if plan != nil && q.allocationReadRepo != nil {
		_ = q.allocationReadRepo.Save(ctx, plan)
	}
	return plan, nil
}
