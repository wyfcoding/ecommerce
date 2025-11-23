package application

import (
	"context"
	"ecommerce/internal/logistics_routing/domain/entity"
	"ecommerce/internal/logistics_routing/domain/repository"
	"errors"

	"log/slog"
)

type LogisticsRoutingService struct {
	repo   repository.LogisticsRoutingRepository
	logger *slog.Logger
}

func NewLogisticsRoutingService(repo repository.LogisticsRoutingRepository, logger *slog.Logger) *LogisticsRoutingService {
	return &LogisticsRoutingService{
		repo:   repo,
		logger: logger,
	}
}

// RegisterCarrier 注册配送商
func (s *LogisticsRoutingService) RegisterCarrier(ctx context.Context, carrier *entity.Carrier) error {
	if err := s.repo.SaveCarrier(ctx, carrier); err != nil {
		s.logger.Error("failed to register carrier", "error", err)
		return err
	}
	return nil
}

// OptimizeRoute 优化路由 (Mock logic)
func (s *LogisticsRoutingService) OptimizeRoute(ctx context.Context, orderIDs []uint64) (*entity.OptimizedRoute, error) {
	// In a real system, this would involve complex algorithms or external APIs.
	// Here we mock a simple assignment.

	carriers, err := s.repo.ListCarriers(ctx, true)
	if err != nil {
		return nil, err
	}
	if len(carriers) == 0 {
		return nil, errors.New("no active carriers found")
	}

	// Simple round-robin or random assignment for mock
	carrier := carriers[0]

	var routeOrders []*entity.RouteOrder
	var totalCost int64

	for _, orderID := range orderIDs {
		cost := carrier.BaseCost + 100 // Mock calculation
		routeOrders = append(routeOrders, &entity.RouteOrder{
			OrderID:       orderID,
			CarrierID:     uint64(carrier.ID),
			CarrierName:   carrier.Name,
			EstimatedCost: cost,
			EstimatedTime: carrier.BaseDeliveryTime,
		})
		totalCost += cost
	}

	route := &entity.OptimizedRoute{
		Orders:      routeOrders,
		OrderCount:  int32(len(orderIDs)),
		TotalCost:   totalCost,
		AverageCost: totalCost / int64(len(orderIDs)),
	}

	if err := s.repo.SaveRoute(ctx, route); err != nil {
		s.logger.Error("failed to save optimized route", "error", err)
		return nil, err
	}

	return route, nil
}

// GetRoute 获取路由
func (s *LogisticsRoutingService) GetRoute(ctx context.Context, id uint64) (*entity.OptimizedRoute, error) {
	return s.repo.GetRoute(ctx, id)
}

// ListCarriers 获取配送商列表
func (s *LogisticsRoutingService) ListCarriers(ctx context.Context) ([]*entity.Carrier, error) {
	return s.repo.ListCarriers(ctx, false)
}
