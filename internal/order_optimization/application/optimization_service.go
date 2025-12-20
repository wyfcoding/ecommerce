package application

import (
	"context"

	"github.com/wyfcoding/ecommerce/internal/order_optimization/domain"
)

// OrderOptimizationService acts as a facade for order optimization operations.
type OrderOptimizationService struct {
	manager *OptimizationManager
	query   *OptimizationQuery
}

// NewOrderOptimizationService creates a new OrderOptimizationService facade.
func NewOrderOptimizationService(manager *OptimizationManager, query *OptimizationQuery) *OrderOptimizationService {
	return &OrderOptimizationService{
		manager: manager,
		query:   query,
	}
}

// --- Write Operations (Delegated to Manager) ---

func (s *OrderOptimizationService) MergeOrders(ctx context.Context, userID uint64, orderIDs []uint64) (*domain.MergedOrder, error) {
	return s.manager.MergeOrders(ctx, userID, orderIDs)
}

func (s *OrderOptimizationService) SplitOrder(ctx context.Context, orderID uint64) ([]*domain.SplitOrder, error) {
	return s.manager.SplitOrder(ctx, orderID)
}

func (s *OrderOptimizationService) AllocateWarehouse(ctx context.Context, orderID uint64) (*domain.WarehouseAllocationPlan, error) {
	return s.manager.AllocateWarehouse(ctx, orderID)
}

// --- Read Operations (Delegated to Query) ---

func (s *OrderOptimizationService) GetMergedOrder(ctx context.Context, id uint64) (*domain.MergedOrder, error) {
	return s.query.GetMergedOrder(ctx, id)
}

func (s *OrderOptimizationService) ListSplitOrders(ctx context.Context, originalOrderID uint64) ([]*domain.SplitOrder, error) {
	return s.query.ListSplitOrders(ctx, originalOrderID)
}

func (s *OrderOptimizationService) GetAllocationPlan(ctx context.Context, orderID uint64) (*domain.WarehouseAllocationPlan, error) {
	return s.query.GetAllocationPlan(ctx, orderID)
}
