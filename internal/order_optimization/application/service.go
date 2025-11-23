package application

import (
	"context"
	"ecommerce/internal/order_optimization/domain/entity"
	"ecommerce/internal/order_optimization/domain/repository"

	"log/slog"
)

type OrderOptimizationService struct {
	repo   repository.OrderOptimizationRepository
	logger *slog.Logger
}

func NewOrderOptimizationService(repo repository.OrderOptimizationRepository, logger *slog.Logger) *OrderOptimizationService {
	return &OrderOptimizationService{
		repo:   repo,
		logger: logger,
	}
}

// MergeOrders 合并订单 (Mock logic)
func (s *OrderOptimizationService) MergeOrders(ctx context.Context, userID uint64, orderIDs []uint64) (*entity.MergedOrder, error) {
	// In a real system, we would fetch original orders, validate they belong to user and can be merged.
	// Here we mock the merge.

	mergedOrder := &entity.MergedOrder{
		UserID:           userID,
		OriginalOrderIDs: orderIDs,
		Items: []*entity.OrderItem{
			{ProductID: 1, Quantity: 2, Price: 100}, // Mock items
		},
		TotalAmount:    200,
		DiscountAmount: 0,
		FinalAmount:    200,
		Status:         "merged",
		ShippingAddress: entity.ShippingAddress{
			Name: "Mock User",
			City: "Mock City",
		},
	}

	if err := s.repo.SaveMergedOrder(ctx, mergedOrder); err != nil {
		s.logger.Error("failed to save merged order", "error", err)
		return nil, err
	}

	return mergedOrder, nil
}

// SplitOrder 拆分订单 (Mock logic)
func (s *OrderOptimizationService) SplitOrder(ctx context.Context, orderID uint64) ([]*entity.SplitOrder, error) {
	// Mock splitting logic: split into 2 orders
	split1 := &entity.SplitOrder{
		OriginalOrderID: orderID,
		SplitIndex:      1,
		Items: []*entity.OrderItem{
			{ProductID: 1, Quantity: 1, Price: 100},
		},
		Amount:      100,
		WarehouseID: 1,
		Status:      "pending",
	}

	split2 := &entity.SplitOrder{
		OriginalOrderID: orderID,
		SplitIndex:      2,
		Items: []*entity.OrderItem{
			{ProductID: 2, Quantity: 1, Price: 100},
		},
		Amount:      100,
		WarehouseID: 2,
		Status:      "pending",
	}

	if err := s.repo.SaveSplitOrder(ctx, split1); err != nil {
		return nil, err
	}
	if err := s.repo.SaveSplitOrder(ctx, split2); err != nil {
		return nil, err
	}

	return []*entity.SplitOrder{split1, split2}, nil
}

// AllocateWarehouse 分配仓库 (Mock logic)
func (s *OrderOptimizationService) AllocateWarehouse(ctx context.Context, orderID uint64) (*entity.WarehouseAllocationPlan, error) {
	plan := &entity.WarehouseAllocationPlan{
		OrderID: orderID,
		Allocations: []*entity.WarehouseAllocation{
			{ProductID: 1, Quantity: 1, WarehouseID: 1, Distance: 10.5},
		},
	}

	if err := s.repo.SaveAllocationPlan(ctx, plan); err != nil {
		return nil, err
	}

	return plan, nil
}
