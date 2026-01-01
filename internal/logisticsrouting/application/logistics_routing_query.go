package application

import (
	"context"

	"github.com/wyfcoding/ecommerce/internal/logisticsrouting/domain"
)

// LogisticsRoutingQuery 处理物流路由的读操作。
type LogisticsRoutingQuery struct {
	repo domain.LogisticsRoutingRepository
}

// NewLogisticsRoutingQuery creates a new LogisticsRoutingQuery instance.
func NewLogisticsRoutingQuery(repo domain.LogisticsRoutingRepository) *LogisticsRoutingQuery {
	return &LogisticsRoutingQuery{
		repo: repo,
	}
}

// GetRoute 获取指定ID的优化路由详情。
func (q *LogisticsRoutingQuery) GetRoute(ctx context.Context, id uint64) (*domain.OptimizedRoute, error) {
	return q.repo.GetRoute(ctx, id)
}

// ListCarriers 获取配送商列表。
func (q *LogisticsRoutingQuery) ListCarriers(ctx context.Context) ([]*domain.Carrier, error) {
	return q.repo.ListCarriers(ctx, false)
}
