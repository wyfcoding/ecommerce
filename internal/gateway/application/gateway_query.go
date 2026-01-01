package application

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"

	"github.com/wyfcoding/ecommerce/internal/gateway/domain"
	"github.com/wyfcoding/pkg/algorithm"
)

// GatewayQuery 处理所有 API 网关相关的查询操作（Queries）。
type GatewayQuery struct {
	repo     domain.GatewayRepository
	logger   *slog.Logger
	hashRing *algorithm.ConsistentHash
}

// NewGatewayQuery 构造函数。
func NewGatewayQuery(repo domain.GatewayRepository, logger *slog.Logger, hashRing *algorithm.ConsistentHash) *GatewayQuery {
	return &GatewayQuery{
		repo:     repo,
		logger:   logger,
		hashRing: hashRing,
	}
}

func (q *GatewayQuery) GetRoute(ctx context.Context, id uint64) (*domain.Route, error) {
	return q.repo.GetRoute(ctx, id)
}

func (q *GatewayQuery) ListRoutes(ctx context.Context, page, pageSize int) ([]*domain.Route, int64, error) {
	offset := (page - 1) * pageSize
	return q.repo.ListRoutes(ctx, offset, pageSize)
}

func (q *GatewayQuery) ListRateLimitRules(ctx context.Context, page, pageSize int) ([]*domain.RateLimitRule, int64, error) {
	offset := (page - 1) * pageSize
	return q.repo.ListRateLimitRules(ctx, offset, pageSize)
}

func (q *GatewayQuery) DispatchByUserID(ctx context.Context, userID uint64, path string) (string, error) {
	key := strconv.FormatUint(userID, 10)
	node := q.hashRing.Get(key)

	if node == "" {
		q.logger.Warn("no backend node available in hash ring", "user_id", userID, "path", path)
		return "", fmt.Errorf("no available backend nodes")
	}

	q.logger.Debug("request dispatched via consistent hash", "user_id", userID, "node", node)
	return node, nil
}
