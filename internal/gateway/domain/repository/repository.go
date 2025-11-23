package repository

import (
	"context"
	"ecommerce/internal/gateway/domain/entity"
)

// GatewayRepository 网关仓储接口
type GatewayRepository interface {
	// 路由管理
	SaveRoute(ctx context.Context, route *entity.Route) error
	GetRoute(ctx context.Context, id uint64) (*entity.Route, error)
	GetRouteByPath(ctx context.Context, path, method string) (*entity.Route, error)
	ListRoutes(ctx context.Context, offset, limit int) ([]*entity.Route, int64, error)
	DeleteRoute(ctx context.Context, id uint64) error

	// 限流规则管理
	SaveRateLimitRule(ctx context.Context, rule *entity.RateLimitRule) error
	GetRateLimitRule(ctx context.Context, id uint64) (*entity.RateLimitRule, error)
	ListRateLimitRules(ctx context.Context, offset, limit int) ([]*entity.RateLimitRule, int64, error)
	DeleteRateLimitRule(ctx context.Context, id uint64) error

	// 日志管理
	SaveAPILog(ctx context.Context, log *entity.APILog) error
}
