package application

import (
	"context"
	"ecommerce/internal/gateway/domain/entity"
	"ecommerce/internal/gateway/domain/repository"

	"log/slog"
)

type GatewayService struct {
	repo   repository.GatewayRepository
	logger *slog.Logger
}

func NewGatewayService(repo repository.GatewayRepository, logger *slog.Logger) *GatewayService {
	return &GatewayService{
		repo:   repo,
		logger: logger,
	}
}

// RegisterRoute 注册路由
func (s *GatewayService) RegisterRoute(ctx context.Context, path, method, service, backend string, timeout, retries int32, description string) (*entity.Route, error) {
	route := entity.NewRoute(path, method, service, backend, timeout, retries, description)
	if err := s.repo.SaveRoute(ctx, route); err != nil {
		s.logger.Error("failed to save route", "error", err)
		return nil, err
	}
	return route, nil
}

// GetRoute 获取路由
func (s *GatewayService) GetRoute(ctx context.Context, id uint64) (*entity.Route, error) {
	return s.repo.GetRoute(ctx, id)
}

// ListRoutes 获取路由列表
func (s *GatewayService) ListRoutes(ctx context.Context, page, pageSize int) ([]*entity.Route, int64, error) {
	offset := (page - 1) * pageSize
	return s.repo.ListRoutes(ctx, offset, pageSize)
}

// DeleteRoute 删除路由
func (s *GatewayService) DeleteRoute(ctx context.Context, id uint64) error {
	return s.repo.DeleteRoute(ctx, id)
}

// AddRateLimitRule 添加限流规则
func (s *GatewayService) AddRateLimitRule(ctx context.Context, name, path, method string, limit, window int32, description string) (*entity.RateLimitRule, error) {
	rule := entity.NewRateLimitRule(name, path, method, limit, window, description)
	if err := s.repo.SaveRateLimitRule(ctx, rule); err != nil {
		s.logger.Error("failed to save rate limit rule", "error", err)
		return nil, err
	}
	return rule, nil
}

// ListRateLimitRules 获取限流规则列表
func (s *GatewayService) ListRateLimitRules(ctx context.Context, page, pageSize int) ([]*entity.RateLimitRule, int64, error) {
	offset := (page - 1) * pageSize
	return s.repo.ListRateLimitRules(ctx, offset, pageSize)
}

// DeleteRateLimitRule 删除限流规则
func (s *GatewayService) DeleteRateLimitRule(ctx context.Context, id uint64) error {
	return s.repo.DeleteRateLimitRule(ctx, id)
}

// LogRequest 记录请求日志
func (s *GatewayService) LogRequest(ctx context.Context, log *entity.APILog) error {
	// In a real high-throughput scenario, this might be async or batched
	return s.repo.SaveAPILog(ctx, log)
}
