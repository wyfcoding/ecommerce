package repository

import (
	"context"
	"github.com/wyfcoding/ecommerce/internal/logistics_routing/domain/entity" // 导入物流路由领域的实体定义。
)

// LogisticsRoutingRepository 是物流路由模块的仓储接口。
// 它定义了对配送商、优化路由和路由统计实体进行数据持久化操作的契约。
// 仓储接口属于领域层，旨在将领域逻辑与数据存储的实现细节解耦。
type LogisticsRoutingRepository interface {
	// --- 配送商 (Carrier methods) ---

	// SaveCarrier 将配送商实体保存到数据存储中。
	// 如果配送商已存在，则更新；如果不存在，则创建。
	// ctx: 上下文。
	// carrier: 待保存的配送商实体。
	SaveCarrier(ctx context.Context, carrier *entity.Carrier) error
	// GetCarrier 根据ID获取配送商实体。
	GetCarrier(ctx context.Context, id uint64) (*entity.Carrier, error)
	// ListCarriers 列出所有配送商实体。
	// activeOnly: 布尔值，如果为true，则只列出活跃的配送商。
	ListCarriers(ctx context.Context, activeOnly bool) ([]*entity.Carrier, error)

	// --- 路由 (OptimizedRoute methods) ---

	// SaveRoute 将优化路由实体保存到数据存储中。
	SaveRoute(ctx context.Context, route *entity.OptimizedRoute) error
	// GetRoute 根据ID获取优化路由实体。
	GetRoute(ctx context.Context, id uint64) (*entity.OptimizedRoute, error)

	// --- 统计 (Statistics methods) ---

	// SaveStatistics 将路由统计实体保存到数据存储中。
	SaveStatistics(ctx context.Context, stats *entity.RoutingStatistics) error
}
