package application

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"github.com/wyfcoding/ecommerce/internal/gateway/domain"
	"github.com/wyfcoding/pkg/algorithm"
)

// GatewayService 结构体定义了API网关相关的应用服务。
// 引入一致性哈希，用于在分布式场景下实现稳定的请求分发（会话粘滞）。
type GatewayService struct {
	repo     domain.GatewayRepository
	logger   *slog.Logger
	hashRing *algorithm.ConsistentHash
}

// NewGatewayService 创建并返回一个新的 GatewayService 实例。
func NewGatewayService(repo domain.GatewayRepository, logger *slog.Logger) *GatewayService {
	return &GatewayService{
		repo:     repo,
		logger:   logger,
		hashRing: algorithm.NewConsistentHash(100, nil), // 100个虚拟节点以平衡分布
	}
}

// RegisterRoute 注册一个新的API路由规则。
func (s *GatewayService) RegisterRoute(ctx context.Context, path, method, service, backend string, timeout, retries int32, description string) (*domain.Route, error) {
	route := domain.NewRoute(path, method, service, backend, timeout, retries, description)
	if err := s.repo.SaveRoute(ctx, route); err != nil {
		s.logger.Error("failed to save route", "error", err)
		return nil, err
	}

	// 动态维护哈希环：当有新的后端节点（Backend）注册时，加入环中
	// 实际场景中，后端地址可能是以逗号分隔的多个实例地址
	backends := strings.Split(backend, ",")
	for _, b := range backends {
		if b != "" {
			s.hashRing.Add(b)
		}
	}

	return route, nil
}

// DispatchByUserID 根据用户ID使用一致性哈希分发请求。
// 这保证了同一个用户的请求始终落到同一个后端节点，有利于本地缓存利用和WebSocket连接管理。
func (s *GatewayService) DispatchByUserID(ctx context.Context, userID uint64, path string) (string, error) {
	// 获取路由信息
	// 简化版：这里假设可以通过路径找到对应的服务节点列表
	// 实际开发中，这里应查询注册中心或本地缓存的后端实例列表
	key := strconv.FormatUint(userID, 10)
	node := s.hashRing.Get(key)

	if node == "" {
		s.logger.Warn("no backend node available in hash ring", "user_id", userID, "path", path)
		return "", fmt.Errorf("no available backend nodes")
	}

	s.logger.Debug("request dispatched via consistent hash", "user_id", userID, "node", node)
	return node, nil
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
