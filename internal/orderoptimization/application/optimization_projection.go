// 生成摘要：新增订单优化读模型投影服务，消费事件后刷新 Redis/ES 读侧。
package application

import (
	"context"
	"log/slog"

	"github.com/wyfcoding/ecommerce/internal/orderoptimization/domain"
)

// OptimizationProjectionService 负责将订单优化事件投影到读模型。
type OptimizationProjectionService struct {
	repo               domain.OrderOptimizationRepository
	mergedReadRepo     domain.MergedOrderReadRepository
	splitReadRepo      domain.SplitOrderReadRepository
	allocationReadRepo domain.AllocationPlanReadRepository
	splitSearchRepo    domain.SplitOrderSearchRepository
	logger             *slog.Logger
}

// NewOptimizationProjectionService 创建投影服务。
func NewOptimizationProjectionService(
	repo domain.OrderOptimizationRepository,
	mergedReadRepo domain.MergedOrderReadRepository,
	splitReadRepo domain.SplitOrderReadRepository,
	allocationReadRepo domain.AllocationPlanReadRepository,
	splitSearchRepo domain.SplitOrderSearchRepository,
	logger *slog.Logger,
) *OptimizationProjectionService {
	return &OptimizationProjectionService{
		repo:               repo,
		mergedReadRepo:     mergedReadRepo,
		splitReadRepo:      splitReadRepo,
		allocationReadRepo: allocationReadRepo,
		splitSearchRepo:    splitSearchRepo,
		logger:             logger,
	}
}

func (s *OptimizationProjectionService) OnOrderMerged(ctx context.Context, event *domain.OrderMergedEvent) error {
	if event == nil {
		return nil
	}
	return s.refreshMergedOrder(ctx, event.MergedOrderID)
}

func (s *OptimizationProjectionService) OnOrderSplit(ctx context.Context, event *domain.OrderSplitEvent) error {
	if event == nil {
		return nil
	}
	return s.refreshSplitOrders(ctx, event.OriginalOrderID)
}

func (s *OptimizationProjectionService) OnAllocationPlanCreated(ctx context.Context, event *domain.OrderAllocationPlanCreatedEvent) error {
	if event == nil {
		return nil
	}
	return s.refreshAllocationPlan(ctx, event.OrderID)
}

func (s *OptimizationProjectionService) refreshMergedOrder(ctx context.Context, mergedOrderID uint64) error {
	if s.mergedReadRepo == nil {
		return nil
	}
	order, err := s.repo.GetMergedOrder(ctx, mergedOrderID)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to load merged order for projection", "merged_order_id", mergedOrderID, "error", err)
		return err
	}
	if order == nil {
		_ = s.mergedReadRepo.Delete(ctx, mergedOrderID)
		return nil
	}
	if err := s.mergedReadRepo.Save(ctx, order); err != nil {
		s.logger.ErrorContext(ctx, "failed to save merged order cache", "merged_order_id", mergedOrderID, "error", err)
		return err
	}
	return nil
}

func (s *OptimizationProjectionService) refreshSplitOrders(ctx context.Context, originalOrderID uint64) error {
	if s.splitReadRepo == nil && s.splitSearchRepo == nil {
		return nil
	}
	orders, err := s.repo.ListSplitOrders(ctx, originalOrderID)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to load split orders for projection", "original_order_id", originalOrderID, "error", err)
		return err
	}
	if s.splitReadRepo != nil {
		if err := s.splitReadRepo.Save(ctx, originalOrderID, orders); err != nil {
			s.logger.ErrorContext(ctx, "failed to save split orders cache", "original_order_id", originalOrderID, "error", err)
			return err
		}
	}
	if s.splitSearchRepo != nil {
		for _, order := range orders {
			if err := s.splitSearchRepo.Index(ctx, order); err != nil {
				s.logger.ErrorContext(ctx, "failed to index split order", "split_order_id", order.ID, "error", err)
				return err
			}
		}
	}
	return nil
}

func (s *OptimizationProjectionService) refreshAllocationPlan(ctx context.Context, orderID uint64) error {
	if s.allocationReadRepo == nil {
		return nil
	}
	plan, err := s.repo.GetAllocationPlan(ctx, orderID)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to load allocation plan for projection", "order_id", orderID, "error", err)
		return err
	}
	if plan == nil {
		_ = s.allocationReadRepo.DeleteByOrderID(ctx, orderID)
		return nil
	}
	if err := s.allocationReadRepo.Save(ctx, plan); err != nil {
		s.logger.ErrorContext(ctx, "failed to save allocation plan cache", "order_id", orderID, "error", err)
		return err
	}
	return nil
}
