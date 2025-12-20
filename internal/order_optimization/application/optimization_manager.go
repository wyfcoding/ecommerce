package application

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/wyfcoding/ecommerce/internal/order_optimization/domain"
)

// OptimizationManager handles write operations for order optimization.
type OptimizationManager struct {
	repo   domain.OrderOptimizationRepository
	logger *slog.Logger
}

// NewOptimizationManager creates a new OptimizationManager instance.
func NewOptimizationManager(repo domain.OrderOptimizationRepository, logger *slog.Logger) *OptimizationManager {
	return &OptimizationManager{
		repo:   repo,
		logger: logger,
	}
}

// MergeOrders 合并订单。
func (m *OptimizationManager) MergeOrders(ctx context.Context, userID uint64, orderIDs []uint64) (*domain.MergedOrder, error) {
	// Mock merge
	mergedOrder := &domain.MergedOrder{
		UserID:           userID,
		OriginalOrderIDs: orderIDs,
		Items: []*domain.OrderItem{
			{ProductID: 1, Quantity: 2, Price: 100},
		},
		TotalAmount:    200,
		DiscountAmount: 0,
		FinalAmount:    200,
		Status:         "merged",
		ShippingAddress: domain.ShippingAddress{
			Name: "Mock User",
			City: "Mock City",
		},
	}

	if err := m.repo.SaveMergedOrder(ctx, mergedOrder); err != nil {
		m.logger.Error("failed to save merged order", "error", err)
		return nil, err
	}

	return mergedOrder, nil
}

// SplitOrder 拆分订单。
func (m *OptimizationManager) SplitOrder(ctx context.Context, orderID uint64) ([]*domain.SplitOrder, error) {
	// Mock splitting logic
	split1 := &domain.SplitOrder{
		OriginalOrderID: orderID,
		SplitIndex:      1,
		Items: []*domain.OrderItem{
			{ProductID: 1, Quantity: 1, Price: 100},
		},
		Amount:      100,
		WarehouseID: 1,
		Status:      "pending",
	}

	split2 := &domain.SplitOrder{
		OriginalOrderID: orderID,
		SplitIndex:      2,
		Items: []*domain.OrderItem{
			{ProductID: 2, Quantity: 1, Price: 100},
		},
		Amount:      100,
		WarehouseID: 2,
		Status:      "pending",
	}

	if err := m.repo.SaveSplitOrder(ctx, split1); err != nil {
		m.logger.Error(fmt.Sprintf("failed to save split order %d-1", orderID), "error", err)
		return nil, err
	}
	if err := m.repo.SaveSplitOrder(ctx, split2); err != nil {
		m.logger.Error(fmt.Sprintf("failed to save split order %d-2", orderID), "error", err)
		return nil, err
	}

	return []*domain.SplitOrder{split1, split2}, nil
}

// AllocateWarehouse 分配仓库。
func (m *OptimizationManager) AllocateWarehouse(ctx context.Context, orderID uint64) (*domain.WarehouseAllocationPlan, error) {
	// Mock logic
	plan := &domain.WarehouseAllocationPlan{
		OrderID: orderID,
		Allocations: []*domain.WarehouseAllocation{
			{ProductID: 1, Quantity: 1, WarehouseID: 1, Distance: 10.5},
		},
	}

	if err := m.repo.SaveAllocationPlan(ctx, plan); err != nil {
		m.logger.Error("failed to save allocation plan", "error", err)
		return nil, err
	}

	return plan, nil
}
