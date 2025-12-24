package application

import (
	"context"
	"errors"
	"log/slog"

	"github.com/wyfcoding/ecommerce/internal/logisticsrouting/domain"
)

// LogisticsRoutingManager 处理物流路由的写操作。
type LogisticsRoutingManager struct {
	repo   domain.LogisticsRoutingRepository
	logger *slog.Logger
}

// NewLogisticsRoutingManager creates a new LogisticsRoutingManager instance.
func NewLogisticsRoutingManager(repo domain.LogisticsRoutingRepository, logger *slog.Logger) *LogisticsRoutingManager {
	return &LogisticsRoutingManager{
		repo:   repo,
		logger: logger,
	}
}

// RegisterCarrier 注册一个新的配送商。
func (m *LogisticsRoutingManager) RegisterCarrier(ctx context.Context, carrier *domain.Carrier) error {
	if err := m.repo.SaveCarrier(ctx, carrier); err != nil {
		m.logger.ErrorContext(ctx, "failed to register carrier", "carrier", carrier.Name, "error", err)
		return err
	}
	m.logger.InfoContext(ctx, "carrier registered successfully", "carrier_id", carrier.ID, "name", carrier.Name)
	return nil
}

// OptimizeRoute 优化配送路线。
func (m *LogisticsRoutingManager) OptimizeRoute(ctx context.Context, orderIDs []uint64) (*domain.OptimizedRoute, error) {
	carriers, err := m.repo.ListCarriers(ctx, true)
	if err != nil {
		return nil, err
	}
	if len(carriers) == 0 {
		return nil, errors.New("no active carriers found for route optimization")
	}

	carrier := carriers[0]

	var routeOrders []*domain.RouteOrder
	var totalCost int64

	for _, orderID := range orderIDs {
		cost := carrier.BaseCost + 100
		routeOrders = append(routeOrders, &domain.RouteOrder{
			OrderID:       orderID,
			CarrierID:     uint64(carrier.ID),
			CarrierName:   carrier.Name,
			EstimatedCost: cost,
			EstimatedTime: carrier.BaseDeliveryTime,
		})
		totalCost += cost
	}

	route := &domain.OptimizedRoute{
		Orders:      routeOrders,
		OrderCount:  int32(len(orderIDs)),
		TotalCost:   totalCost,
		AverageCost: totalCost / int64(len(orderIDs)),
	}

	if err := m.repo.SaveRoute(ctx, route); err != nil {
		m.logger.ErrorContext(ctx, "failed to save optimized route", "error", err)
		return nil, err
	}
	m.logger.InfoContext(ctx, "optimized route created", "route_id", route.ID, "order_count", route.OrderCount)

	return route, nil
}
