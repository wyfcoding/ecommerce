package application

import (
	"context"
	"errors" // 导入标准错误处理库。

	// 导入格式化库。
	// 导入时间库。
	"github.com/wyfcoding/ecommerce/internal/logistics_routing/domain/entity"     // 导入物流路由领域的实体定义。
	"github.com/wyfcoding/ecommerce/internal/logistics_routing/domain/repository" // 导入物流路由领域的仓储接口。

	"log/slog" // 导入结构化日志库。
)

// LogisticsRoutingService 结构体定义了物流路由相关的应用服务。
// 它协调领域层和基础设施层，处理配送商的注册、配送路线的优化和管理等业务逻辑。
type LogisticsRoutingService struct {
	repo   repository.LogisticsRoutingRepository // 依赖LogisticsRoutingRepository接口，用于数据持久化操作。
	logger *slog.Logger                          // 日志记录器，用于记录服务运行时的信息和错误。
}

// NewLogisticsRoutingService 创建并返回一个新的 LogisticsRoutingService 实例。
func NewLogisticsRoutingService(repo repository.LogisticsRoutingRepository, logger *slog.Logger) *LogisticsRoutingService {
	return &LogisticsRoutingService{
		repo:   repo,
		logger: logger,
	}
}

// RegisterCarrier 注册一个新的配送商。
// ctx: 上下文。
// carrier: 待注册的Carrier实体。
// 返回可能发生的错误。
func (s *LogisticsRoutingService) RegisterCarrier(ctx context.Context, carrier *entity.Carrier) error {
	if err := s.repo.SaveCarrier(ctx, carrier); err != nil {
		s.logger.Error("failed to register carrier", "error", err)
		return err
	}
	return nil
}

// OptimizeRoute 优化配送路线。
// ctx: 上下文。
// orderIDs: 待配送的订单ID列表。
// 返回计算出的OptimizedRoute实体和可能发生的错误。
func (s *LogisticsRoutingService) OptimizeRoute(ctx context.Context, orderIDs []uint64) (*entity.OptimizedRoute, error) {
	// TODO: 在实际系统中，这会涉及复杂的算法（如旅行商问题、车辆路径问题）或调用外部物流API。
	// 当前实现是一个简化的模拟：为所有订单分配给第一个活跃的配送商。

	// 1. 获取所有活跃的配送商。
	carriers, err := s.repo.ListCarriers(ctx, true) // true表示只列出活跃的配送商。
	if err != nil {
		return nil, err
	}
	if len(carriers) == 0 {
		return nil, errors.New("no active carriers found for route optimization")
	}

	// 2. 简单地选择第一个活跃的配送商（模拟分配）。
	// 实际情况可能需要更复杂的选择逻辑（例如，基于承运商能力、成本、区域）。
	carrier := carriers[0]

	var routeOrders []*entity.RouteOrder
	var totalCost int64

	// 为每个订单创建一个路由订单记录，并模拟计算成本和时间。
	for _, orderID := range orderIDs {
		cost := carrier.BaseCost + 100 // 模拟计算配送成本。
		routeOrders = append(routeOrders, &entity.RouteOrder{
			OrderID:       orderID,
			CarrierID:     uint64(carrier.ID),
			CarrierName:   carrier.Name,
			EstimatedCost: cost,
			EstimatedTime: carrier.BaseDeliveryTime, // 使用配送商的基础配送时间。
		})
		totalCost += cost
	}

	// 创建优化后的路由实体。
	route := &entity.OptimizedRoute{
		Orders:      routeOrders,
		OrderCount:  int32(len(orderIDs)),
		TotalCost:   totalCost,
		AverageCost: totalCost / int64(len(orderIDs)), // 计算平均成本。
	}

	// 通过仓储接口保存优化后的路由。
	if err := s.repo.SaveRoute(ctx, route); err != nil {
		s.logger.Error("failed to save optimized route", "error", err)
		return nil, err
	}

	return route, nil
}

// GetRoute 获取指定ID的优化路由详情。
// ctx: 上下文。
// id: 优化路由ID。
// 返回OptimizedRoute实体和可能发生的错误。
func (s *LogisticsRoutingService) GetRoute(ctx context.Context, id uint64) (*entity.OptimizedRoute, error) {
	return s.repo.GetRoute(ctx, id)
}

// ListCarriers 获取配送商列表。
// ctx: 上下文。
// 返回Carrier列表和可能发生的错误。
func (s *LogisticsRoutingService) ListCarriers(ctx context.Context) ([]*entity.Carrier, error) {
	return s.repo.ListCarriers(ctx, false) // false表示列出所有配送商（包括非活跃的）。
}
