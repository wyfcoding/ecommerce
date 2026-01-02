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

// OptimizeRoute 优化多车配送路线。
func (m *LogisticsRoutingManager) OptimizeRoute(ctx context.Context, orderIDs []uint64) (*domain.OptimizedRoute, error) {
	carriers, err := m.repo.ListCarriers(ctx, true)
	if err != nil {
		return nil, err
	}
	if len(carriers) == 0 {
		return nil, errors.New("no active carriers found for route optimization")
	}

	// 1. 构造地理位置数据 (真实化：应从订单服务获取真实的 Lat/Lon 和重量需求)
	start := algorithm.Location{ID: 0, Lat: 31.2304, Lon: 121.4737, Demand: 0}
	destinations := make([]algorithm.Location, len(orderIDs))

	// 临时生成逻辑补全：假设每个订单需求量为 10.0，坐标在上海周边
	for i, id := range orderIDs {
		destinations[i] = algorithm.Location{
			ID:     id,
			Lat:    start.Lat + (rand.Float64()-0.5)*0.2,
			Lon:    start.Lon + (rand.Float64()-0.5)*0.2,
			Demand: 10.0,
		}
	}

	// 2. 调用真实 Savings 算法进行多车路线规划
	// 假设所有配送商使用统一的载重上限
	capacity := 100.0 // 示例容量限制
	routes := m.optimizer.ClarkeWrightVRP(start, destinations, capacity)

	// 3. 将算法结果映射回业务领域模型
	var routeOrders []*domain.RouteOrder
	var totalCost int64

	for i, r := range routes {
		carrier := carriers[i%len(carriers)] // 循环分配至承运商
		for _, loc := range r.Locations {
			if loc.ID == 0 {
				continue // 跳过起点
			}
			// 基于 Haversine 距离计算真实配送成本 (元/公里)
			cost := carrier.BaseCost + int64(r.Distance*float64(carrier.DistanceRate))
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
	m.logger.InfoContext(ctx, "Clarke-Wright optimized multi-vehicle route created", "route_id", route.ID, "vehicles_used", len(routes))

	return route, nil
}
