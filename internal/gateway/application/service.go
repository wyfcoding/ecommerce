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
type GatewayService struct {
	repo     domain.GatewayRepository
	logger   *slog.Logger
	hashRing *algorithm.ConsistentHash
}

// SyncRouteRequest 声明式同步请求
type SyncRouteRequest struct {
	ExternalID string // 外部唯一标识 (如 K8s UID)
	Name       string
	Path       string
	Method     string
	Backend    string
	Source     string // 配置来源: K8S_CRD, API, ETCD
}

// NewGatewayService 创建并返回一个新的 GatewayService 实例。
func NewGatewayService(repo domain.GatewayRepository, logger *slog.Logger) *GatewayService {
	return &GatewayService{
		repo:     repo,
		logger:   logger,
		hashRing: algorithm.NewConsistentHash(100, nil),
	}
}

// RegisterRoute 注册一个新的API路由规则 (兼容 API 模式)。
func (s *GatewayService) RegisterRoute(ctx context.Context, path, method, service, backend string, timeout, retries int32, description string) (*domain.Route, error) {
	route := domain.NewRoute(path, method, service, backend, timeout, retries, description)
	if err := s.repo.SaveRoute(ctx, route); err != nil {
		s.logger.Error("failed to save route", "error", err)
		return nil, err
	}

	s.updateHashRing(backend)
	return route, nil
}

// SyncRoute 声明式同步：实现“最终状态对齐” (适配 K8s CRD 模式)
func (s *GatewayService) SyncRoute(ctx context.Context, req SyncRouteRequest) (*domain.Route, error) {
	existing, err := s.repo.GetRouteByExternalID(ctx, req.ExternalID)
	if err != nil {
		return nil, err
	}

	var route *domain.Route
	if existing != nil {
		existing.Path = req.Path
		existing.Method = req.Method
		existing.Backend = req.Backend
		existing.Description = fmt.Sprintf("Synced from %s: %s", req.Source, req.Name)
		route = existing
	} else {
		route = domain.NewRoute(req.Path, req.Method, "gateway", req.Backend, 30, 3, req.Name)
		route.ExternalID = req.ExternalID
	}

	if err := s.repo.SaveRoute(ctx, route); err != nil {
		return nil, err
	}

	s.updateHashRing(req.Backend)
	return route, nil
}

// DeleteRouteByExternalID 根据外部 ID 删除路由
func (s *GatewayService) DeleteRouteByExternalID(ctx context.Context, externalID string) error {
	route, err := s.repo.GetRouteByExternalID(ctx, externalID)
	if err != nil {
		return err
	}
	if route != nil {
		return s.repo.DeleteRoute(ctx, uint64(route.ID))
	}
	return nil
}

func (s *GatewayService) updateHashRing(backend string) {
	backends := strings.Split(backend, ",")
	for _, b := range backends {
		if b != "" {
			s.hashRing.Add(b)
		}
	}
}

// DispatchByUserID 根据用户ID使用一致性哈希分发请求。
func (s *GatewayService) DispatchByUserID(ctx context.Context, userID uint64, path string) (string, error) {
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