package application

import (
	"context"

	"github.com/wyfcoding/ecommerce/internal/gateway/domain/entity"     // 导入网关领域的实体定义。
	"github.com/wyfcoding/ecommerce/internal/gateway/domain/repository" // 导入网关领域的仓储接口。

	"log/slog" // 导入结构化日志库。
)

// Gateway 结构体定义了API网关相关的应用服务。
// 它协调领域层和基础设施层，处理路由管理、限流规则配置和API请求日志记录等业务逻辑。
type Gateway struct {
	repo   repository.GatewayRepository // 依赖GatewayRepository接口，用于数据持久化操作。
	logger *slog.Logger                 // 日志记录器，用于记录服务运行时的信息和错误。
}

// NewGateway 创建并返回一个新的 Gateway 实例。
func NewGateway(repo repository.GatewayRepository, logger *slog.Logger) *Gateway {
	return &Gateway{
		repo:   repo,
		logger: logger,
	}
}

// RegisterRoute 注册一个新的API路由规则。
// ctx: 上下文。
// path: 路由路径。
// method: HTTP方法。
// service: 目标服务名称。
// backend: 后端服务地址。
// timeout: 请求超时时间。
// retries: 重试次数。
// description: 路由描述。
// 返回created successfully的Route实体和可能发生的错误。
func (s *Gateway) RegisterRoute(ctx context.Context, path, method, service, backend string, timeout, retries int32, description string) (*entity.Route, error) {
	route := entity.NewRoute(path, method, service, backend, timeout, retries, description) // 创建Route实体。
	// 通过仓储接口保存路由。
	if err := s.repo.SaveRoute(ctx, route); err != nil {
		s.logger.Error("failed to save route", "error", err)
		return nil, err
	}
	return route, nil
}

// GetRoute 获取指定ID的路由规则。
// ctx: 上下文。
// id: 路由ID。
// 返回Route实体和可能发生的错误。
func (s *Gateway) GetRoute(ctx context.Context, id uint64) (*entity.Route, error) {
	return s.repo.GetRoute(ctx, id)
}

// ListRoutes 获取路由规则列表。
// ctx: 上下文。
// page, pageSize: 分页参数。
// 返回路由列表、总数和可能发生的错误。
func (s *Gateway) ListRoutes(ctx context.Context, page, pageSize int) ([]*entity.Route, int64, error) {
	offset := (page - 1) * pageSize
	return s.repo.ListRoutes(ctx, offset, pageSize)
}

// DeleteRoute 删除指定ID的路由规则。
// ctx: 上下文。
// id: 路由ID。
// 返回可能发生的错误。
func (s *Gateway) DeleteRoute(ctx context.Context, id uint64) error {
	return s.repo.DeleteRoute(ctx, id)
}

// AddRateLimitRule 添加一个新的API限流规则。
// ctx: 上下文。
// name: 规则名称。
// path: 匹配路径。
// method: 匹配HTTP方法。
// limit: 限流次数。
// window: 限流时间窗口。
// description: 规则描述。
// 返回created successfully的RateLimitRule实体和可能发生的错误。
func (s *Gateway) AddRateLimitRule(ctx context.Context, name, path, method string, limit, window int32, description string) (*entity.RateLimitRule, error) {
	rule := entity.NewRateLimitRule(name, path, method, limit, window, description) // 创建RateLimitRule实体。
	// 通过仓储接口保存限流规则。
	if err := s.repo.SaveRateLimitRule(ctx, rule); err != nil {
		s.logger.Error("failed to save rate limit rule", "error", err)
		return nil, err
	}
	return rule, nil
}

// ListRateLimitRules 获取限流规则列表。
// ctx: 上下文。
// page, pageSize: 分页参数。
// 返回限流规则列表、总数和可能发生的错误。
func (s *Gateway) ListRateLimitRules(ctx context.Context, page, pageSize int) ([]*entity.RateLimitRule, int64, error) {
	offset := (page - 1) * pageSize
	return s.repo.ListRateLimitRules(ctx, offset, pageSize)
}

// DeleteRateLimitRule 删除指定ID的限流规则。
// ctx: 上下文。
// id: 限流规则ID。
// 返回可能发生的错误。
func (s *Gateway) DeleteRateLimitRule(ctx context.Context, id uint64) error {
	return s.repo.DeleteRateLimitRule(ctx, id)
}

// LogRequest 记录API请求日志。
// ctx: 上下文。
// log: 待保存的APILog实体。
// 返回可能发生的错误。
func (s *Gateway) LogRequest(ctx context.Context, log *entity.APILog) error {
	// 在实际高吞吐量场景中，此操作可能需要异步或批量处理。
	return s.repo.SaveAPILog(ctx, log)
}
