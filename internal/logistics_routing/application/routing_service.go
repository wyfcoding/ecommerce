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

// NewLogisticsRoutingService creates a new LogisticsRoutingService facade.
func NewLogisticsRoutingService(manager *LogisticsRoutingManager, query *LogisticsRoutingQuery) *LogisticsRoutingService {
	return &LogisticsRoutingService{
		manager: manager,
		query:   query,
	}
}

// --- 写操作（委托给 Manager）---

func (s *LogisticsRoutingService) RegisterCarrier(ctx context.Context, carrier *domain.Carrier) error {
	return s.manager.RegisterCarrier(ctx, carrier)
}

func (s *LogisticsRoutingService) OptimizeRoute(ctx context.Context, orderIDs []uint64) (*domain.OptimizedRoute, error) {
	return s.manager.OptimizeRoute(ctx, orderIDs)
}

// --- 读操作（委托给 Query）---

func (s *LogisticsRoutingService) GetRoute(ctx context.Context, id uint64) (*domain.OptimizedRoute, error) {
	return s.query.GetRoute(ctx, id)
}

func (s *LogisticsRoutingService) ListCarriers(ctx context.Context) ([]*domain.Carrier, error) {
	return s.query.ListCarriers(ctx)
}
