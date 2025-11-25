package repository

import (
	"context"
	"github.com/wyfcoding/ecommerce/internal/logistics_routing/domain/entity"
)

// LogisticsRoutingRepository 物流路由仓储接口
type LogisticsRoutingRepository interface {
	// 配送商
	SaveCarrier(ctx context.Context, carrier *entity.Carrier) error
	GetCarrier(ctx context.Context, id uint64) (*entity.Carrier, error)
	ListCarriers(ctx context.Context, activeOnly bool) ([]*entity.Carrier, error)

	// 路由
	SaveRoute(ctx context.Context, route *entity.OptimizedRoute) error
	GetRoute(ctx context.Context, id uint64) (*entity.OptimizedRoute, error)

	// 统计
	SaveStatistics(ctx context.Context, stats *entity.RoutingStatistics) error
}
