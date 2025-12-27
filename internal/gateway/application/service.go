package application

import (
	"context"
	"log/slog"

	"github.com/wyfcoding/ecommerce/internal/gateway/domain"
)

// GatewayService 结构体定义了API网关相关的应用服务。
type GatewayService struct {
	repo   domain.GatewayRepository
	logger *slog.Logger
}

// NewGatewayService 创建并返回一个新的 GatewayService 实例。
func NewGatewayService(repo domain.GatewayRepository, logger *slog.Logger) *GatewayService {
	return &GatewayService{
		repo:   repo,
		logger: logger,
	}
}

// RegisterRoute 注册一个新的API路由规则。
func (s *GatewayService) RegisterRoute(ctx context.Context, path, method, service, backend string, timeout, retries int32, description string) (*domain.Route, error) {
	route := domain.NewRoute(path, method, service, backend, timeout, retries, description)
	if err := s.repo.SaveRoute(ctx, route); err != nil {
		s.logger.Error("failed to save route", "error", err)
		return nil, err
	}
	return route, nil
}

// GetRoute 获取指定ID的路由规则。
func (s *GatewayService) GetRoute(ctx context.Context, id uint64) (*domain.Route, error) {
	return s.repo.GetRoute(ctx, id)
}

// ListRoutes 获取路由规则列表。
func (s *GatewayService) ListRoutes(ctx context.Context, page, pageSize int) ([]*domain.Route, int64, error) {
	offset := (page - 1) * pageSize
	return s.repo.ListRoutes(ctx, offset, pageSize)
}

// DeleteRoute 删除指定ID的路由规则。
func (s *GatewayService) DeleteRoute(ctx context.Context, id uint64) error {
	return s.repo.DeleteRoute(ctx, id)
}

// AddRateLimitRule 添加一个新的API限流规则。
func (s *GatewayService) AddRateLimitRule(ctx context.Context, name, path, method string, limit, window int32, description string) (*domain.RateLimitRule, error) {
	rule := domain.NewRateLimitRule(name, path, method, limit, window, description)
	if err := s.repo.SaveRateLimitRule(ctx, rule); err != nil {
		s.logger.Error("failed to save rate limit rule", "error", err)
		return nil, err
	}
	return rule, nil
}

// ListRateLimitRules 获取限流规则列表。
func (s *GatewayService) ListRateLimitRules(ctx context.Context, page, pageSize int) ([]*domain.RateLimitRule, int64, error) {
	offset := (page - 1) * pageSize
	return s.repo.ListRateLimitRules(ctx, offset, pageSize)
}

// DeleteRateLimitRule 删除指定ID的限流规则。
func (s *GatewayService) DeleteRateLimitRule(ctx context.Context, id uint64) error {
	return s.repo.DeleteRateLimitRule(ctx, id)
}

// LogRequest 记录API请求日志。
func (s *GatewayService) LogRequest(ctx context.Context, log *domain.APILog) error {
	return s.repo.SaveAPILog(ctx, log)
}
