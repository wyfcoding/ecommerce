package domain

import (
	"context"
)

// LogisticsRoutingRepository 是物流路由模块的仓储接口。
type LogisticsRoutingRepository interface {
	// 配送商
	SaveCarrier(ctx context.Context, carrier *Carrier) error
	GetCarrier(ctx context.Context, id uint64) (*Carrier, error)
	ListCarriers(ctx context.Context, activeOnly bool) ([]*Carrier, error)

	// 路由
	SaveRoute(ctx context.Context, route *OptimizedRoute) error
	GetRoute(ctx context.Context, id uint64) (*OptimizedRoute, error)

	// 统计
	SaveStatistics(ctx context.Context, stats *RoutingStatistics) error
}
