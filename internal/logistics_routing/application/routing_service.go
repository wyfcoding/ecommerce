package application

import (
	"context"

	"github.com/wyfcoding/ecommerce/internal/logistics_routing/domain"
)

// LogisticsRoutingService 作为物流路由操作的门面。
type LogisticsRoutingService struct {
	manager *LogisticsRoutingManager
	query   *LogisticsRoutingQuery
}

// NewLogisticsRoutingService 创建物流路由服务门面实例。
func NewLogisticsRoutingService(manager *LogisticsRoutingManager, query *LogisticsRoutingQuery) *LogisticsRoutingService {
	return &LogisticsRoutingService{
		manager: manager,
		query:   query,
	}
}

// --- 写操作（委托给 Manager）---

// RegisterCarrier 注册一个新的物流承运商信息。
func (s *LogisticsRoutingService) RegisterCarrier(ctx context.Context, carrier *domain.Carrier) error {
	return s.manager.RegisterCarrier(ctx, carrier)
}

// OptimizeRoute 核心算法：为一组订单优化整体配送路由。
func (s *LogisticsRoutingService) OptimizeRoute(ctx context.Context, orderIDs []uint64) (*domain.OptimizedRoute, error) {
	return s.manager.OptimizeRoute(ctx, orderIDs)
}

// --- 读操作（委托给 Query）---

// GetRoute 获取指定ID的已优化路由详情。
func (s *LogisticsRoutingService) GetRoute(ctx context.Context, id uint64) (*domain.OptimizedRoute, error) {
	return s.query.GetRoute(ctx, id)
}

// ListCarriers 获取所有可用的物流承运商列表。
func (s *LogisticsRoutingService) ListCarriers(ctx context.Context) ([]*domain.Carrier, error) {
	return s.query.ListCarriers(ctx)
}
