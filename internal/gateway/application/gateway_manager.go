package application

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/wyfcoding/ecommerce/internal/gateway/domain"
	"github.com/wyfcoding/pkg/algorithm"
)

// GatewayManager 处理所有 API 网关相关的写入操作（Commands）。
type GatewayManager struct {
	repo     domain.GatewayRepository
	logger   *slog.Logger
	hashRing *algorithm.ConsistentHash
}

// NewGatewayManager 构造函数。
func NewGatewayManager(repo domain.GatewayRepository, logger *slog.Logger, hashRing *algorithm.ConsistentHash) *GatewayManager {
	return &GatewayManager{
		repo:     repo,
		logger:   logger,
		hashRing: hashRing,
	}
}

func (m *GatewayManager) RegisterRoute(ctx context.Context, path, method, service, backend string, timeout, retries int32, description string) (*domain.Route, error) {
	route := domain.NewRoute(path, method, service, backend, timeout, retries, description)
	if err := m.repo.SaveRoute(ctx, route); err != nil {
		m.logger.Error("failed to save route", "error", err)
		return nil, err
	}

	m.updateHashRing(backend)
	return route, nil
}

func (m *GatewayManager) SyncRoute(ctx context.Context, req SyncRouteRequest) (*domain.Route, error) {
	existing, err := m.repo.GetRouteByExternalID(ctx, req.ExternalID)
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

	if err := m.repo.SaveRoute(ctx, route); err != nil {
		return nil, err
	}

	m.updateHashRing(req.Backend)
	return route, nil
}

func (m *GatewayManager) DeleteRoute(ctx context.Context, id uint64) error {
	return m.repo.DeleteRoute(ctx, id)
}

func (m *GatewayManager) DeleteRouteByExternalID(ctx context.Context, externalID string) error {
	route, err := m.repo.GetRouteByExternalID(ctx, externalID)
	if err != nil {
		return err
	}
	if route != nil {
		return m.repo.DeleteRoute(ctx, uint64(route.ID))
	}
	return nil
}

func (m *GatewayManager) updateHashRing(backend string) {
	slog.Info("Updating hash ring with backends", "backends", backend)
	for b := range strings.SplitSeq(backend, ",") {
		if b != "" {
			m.hashRing.Add(b)
		}
	}
}

func (m *GatewayManager) AddRateLimitRule(ctx context.Context, name, path, method string, limit, window int32, description string) (*domain.RateLimitRule, error) {
	rule := domain.NewRateLimitRule(name, path, method, limit, window, description)
	if err := m.repo.SaveRateLimitRule(ctx, rule); err != nil {
		m.logger.Error("failed to save rate limit rule", "error", err)
		return nil, err
	}
	return rule, nil
}

func (m *GatewayManager) DeleteRateLimitRule(ctx context.Context, id uint64) error {
	return m.repo.DeleteRateLimitRule(ctx, id)
}

func (m *GatewayManager) LogRequest(ctx context.Context, log *domain.APILog) error {
	return m.repo.SaveAPILog(ctx, log)
}

// SyncRouteRequest 声明式同步请求
type SyncRouteRequest struct {
	ExternalID string
	Name       string
	Path       string
	Method     string
	Backend    string
	Source     string
}
