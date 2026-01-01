package domain

import (
	"context"
)

// GatewayRepository 是API网关模块的仓储接口。
// 它定义了对路由规则、限流规则和API日志实体进行数据持久化操作的契约。
// 仓储接口属于领域层，旨在将领域逻辑与数据存储的实现细节解耦。
type GatewayRepository interface {
	// --- 路由管理 (Route methods) ---

	// SaveRoute 将路由实体保存到数据存储中。
	// 如果路由已存在，则更新；如果不存在，则创建。
	// ctx: 上下文。
	// route: 待保存的路由实体。
	SaveRoute(ctx context.Context, route *Route) error
	// GetRoute 根据ID获取路由实体。
	GetRoute(ctx context.Context, id uint64) (*Route, error)
	// GetRouteByPath 根据路径和HTTP方法获取路由实体。
	GetRouteByPath(ctx context.Context, path, method string) (*Route, error)
	// ListRoutes 列出所有路由实体，支持分页。
	ListRoutes(ctx context.Context, offset, limit int) ([]*Route, int64, error)
	// DeleteRoute 根据ID删除路由实体。
	DeleteRoute(ctx context.Context, id uint64) error
	// GetRouteByExternalID 根据外部 ID 获取路由实体。
	GetRouteByExternalID(ctx context.Context, externalID string) (*Route, error)

	// --- 限流规则管理 (RateLimitRule methods) ---

	// SaveRateLimitRule 将限流规则实体保存到数据存储中。
	SaveRateLimitRule(ctx context.Context, rule *RateLimitRule) error
	// GetRateLimitRule 根据ID获取限流规则实体。
	GetRateLimitRule(ctx context.Context, id uint64) (*RateLimitRule, error)
	// ListRateLimitRules 列出所有限流规则实体，支持分页。
	ListRateLimitRules(ctx context.Context, offset, limit int) ([]*RateLimitRule, int64, error)
	// DeleteRateLimitRule 根据ID删除限流规则实体。
	DeleteRateLimitRule(ctx context.Context, id uint64) error

	// --- 日志管理 (APILog methods) ---

	// SaveAPILog 将API日志实体保存到数据存储中。
	SaveAPILog(ctx context.Context, log *APILog) error
}
