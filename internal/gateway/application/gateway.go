package application

import (
	"context"
	"log/slog"

	"github.com/wyfcoding/ecommerce/internal/gateway/domain"
	"github.com/wyfcoding/pkg/algorithm"
)

// GatewayService API 网关门面服务，整合 Manager 和 Query。
type GatewayService struct {
	manager *GatewayManager
	query   *GatewayQuery
}

// NewGatewayService 构造函数。
func NewGatewayService(repo domain.GatewayRepository, logger *slog.Logger) *GatewayService {
	// 共享一致性哈希环
	hashRing := algorithm.NewConsistentHash(100, nil)
	return &GatewayService{
		manager: NewGatewayManager(repo, logger, hashRing),
		query:   NewGatewayQuery(repo, logger, hashRing),
	}
}

// --- Manager (Writes) ---

func (s *GatewayService) RegisterRoute(ctx context.Context, path, method, service, backend string, timeout, retries int32, description string) (*domain.Route, error) {
	return s.manager.RegisterRoute(ctx, path, method, service, backend, timeout, retries, description)
}

func (s *GatewayService) SyncRoute(ctx context.Context, req SyncRouteRequest) (*domain.Route, error) {
	return s.manager.SyncRoute(ctx, req)
}

func (s *GatewayService) DeleteRoute(ctx context.Context, id uint64) error {
	return s.manager.DeleteRoute(ctx, id)
}

func (s *GatewayService) DeleteRouteByExternalID(ctx context.Context, externalID string) error {
	return s.manager.DeleteRouteByExternalID(ctx, externalID)
}

func (s *GatewayService) AddRateLimitRule(ctx context.Context, name, path, method string, limit, window int32, description string) (*domain.RateLimitRule, error) {
	return s.manager.AddRateLimitRule(ctx, name, path, method, limit, window, description)
}

func (s *GatewayService) DeleteRateLimitRule(ctx context.Context, id uint64) error {
	return s.manager.DeleteRateLimitRule(ctx, id)
}

func (s *GatewayService) LogRequest(ctx context.Context, log *domain.APILog) error {
	return s.manager.LogRequest(ctx, log)
}

// --- Query (Reads) ---

func (s *GatewayService) GetRoute(ctx context.Context, id uint64) (*domain.Route, error) {
	return s.query.GetRoute(ctx, id)
}

func (s *GatewayService) ListRoutes(ctx context.Context, page, pageSize int) ([]*domain.Route, int64, error) {
	return s.query.ListRoutes(ctx, page, pageSize)
}

func (s *GatewayService) ListRateLimitRules(ctx context.Context, page, pageSize int) ([]*domain.RateLimitRule, int64, error) {
	return s.query.ListRateLimitRules(ctx, page, pageSize)
}

func (s *GatewayService) DispatchByUserID(ctx context.Context, userID uint64, path string) (string, error) {
	return s.query.DispatchByUserID(ctx, userID, path)
}
