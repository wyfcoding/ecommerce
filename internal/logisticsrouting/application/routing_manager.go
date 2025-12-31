package application

import (
	"context"
	"errors"
	"log/slog"
	"math/rand"

	"github.com/wyfcoding/ecommerce/internal/logisticsrouting/domain"
	"github.com/wyfcoding/pkg/algorithm"
)

// LogisticsRoutingManager 处理物流路由的写操作。
type LogisticsRoutingManager struct {
	repo      domain.LogisticsRoutingRepository
	optimizer *algorithm.RouteOptimizer
	logger    *slog.Logger
}

// NewLogisticsRoutingManager creates a new LogisticsRoutingManager instance.
func NewLogisticsRoutingManager(repo domain.LogisticsRoutingRepository, logger *slog.Logger) *LogisticsRoutingManager {
	return &LogisticsRoutingManager{
		repo:      repo,
		optimizer: algorithm.NewRouteOptimizer(),
		logger:    logger,
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

	// 1. 构造地理位置数据 (实际应从订单详情中获取 Lat/Lon)
	start := algorithm.Location{ID: 0, Lat: 31.2304, Lon: 121.4737} // 假设从上海中心仓库出发
	destinations := make([]algorithm.Location, len(orderIDs))
	for i, id := range orderIDs {
		destinations[i] = algorithm.Location{
			ID:  id,
			Lat: start.Lat + (rand.Float64()-0.5)*0.5, // 随机生成周边的配送点
			Lon: start.Lon + (rand.Float64()-0.5)*0.5,
		}
	}

	// 2. 调用算法进行多车路线规划 (VRP 简化版)
	// 假设每辆车可以处理的量是有限的，或者我们根据配送商数量进行拆分
	numVehicles := len(carriers)
	if numVehicles > len(orderIDs) {
		numVehicles = len(orderIDs)
	}

	routes := m.optimizer.OptimizeBatchRoutes(start, destinations, numVehicles)

	// 3. 将算法结果映射回业务领域模型
	var routeOrders []*domain.RouteOrder
	var totalCost int64

	for i, r := range routes {
		carrier := carriers[i%len(carriers)] // 简单分配给承运商
		for _, loc := range r.Locations {
			if loc.ID == 0 {
				continue // 跳过起点
			}
			cost := carrier.BaseCost + int64(r.Distance/1000.0*carrier.DistanceRate)
			routeOrders = append(routeOrders, &domain.RouteOrder{
				OrderID:       loc.ID,
				CarrierID:     uint64(carrier.ID),
				CarrierName:   carrier.Name,
				EstimatedCost: cost,
				EstimatedTime: carrier.BaseDeliveryTime,
			})
			totalCost += cost
		}
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
	m.logger.InfoContext(ctx, "multi-vehicle optimized route created", "route_id", route.ID, "order_count", route.OrderCount, "vehicles_used", len(routes))

	return route, nil
}
