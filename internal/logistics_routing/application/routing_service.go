package application

import (
	"context"

	"github.com/wyfcoding/ecommerce/internal/logistics_routing/domain"
)

// LogisticsRoutingService acts as a facade for logistics routing operations.
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

// --- Write Operations (Delegated to Manager) ---

func (s *LogisticsRoutingService) RegisterCarrier(ctx context.Context, carrier *domain.Carrier) error {
	return s.manager.RegisterCarrier(ctx, carrier)
}

func (s *LogisticsRoutingService) OptimizeRoute(ctx context.Context, orderIDs []uint64) (*domain.OptimizedRoute, error) {
	return s.manager.OptimizeRoute(ctx, orderIDs)
}

// --- Read Operations (Delegated to Query) ---

func (s *LogisticsRoutingService) GetRoute(ctx context.Context, id uint64) (*domain.OptimizedRoute, error) {
	return s.query.GetRoute(ctx, id)
}

func (s *LogisticsRoutingService) ListCarriers(ctx context.Context) ([]*domain.Carrier, error) {
	return s.query.ListCarriers(ctx)
}
