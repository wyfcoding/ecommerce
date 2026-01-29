package application

import (
	"context"

	"github.com/wyfcoding/ecommerce/internal/logistics/domain"
)

// LogisticsQuery 负责物流的查询操作。
type LogisticsQuery struct {
	repo domain.LogisticsRepository
}

// NewLogisticsQuery 构造函数。
func NewLogisticsQuery(repo domain.LogisticsRepository) *LogisticsQuery {
	return &LogisticsQuery{repo: repo}
}

func (q *LogisticsQuery) GetLogistics(ctx context.Context, id uint64) (*domain.Logistics, error) {
	return q.repo.GetByID(ctx, id)
}

func (q *LogisticsQuery) GetLogisticsByTrackingNo(ctx context.Context, trackingNo string) (*domain.Logistics, error) {
	return q.repo.GetByTrackingNo(ctx, trackingNo)
}

func (q *LogisticsQuery) GetLogisticsByOrderID(ctx context.Context, orderID uint64) (*domain.Logistics, error) {
	return q.repo.GetByOrderID(ctx, orderID)
}

func (q *LogisticsQuery) ListLogistics(ctx context.Context, page, pageSize int) ([]*domain.Logistics, int64, error) {
	offset := (page - 1) * pageSize
	return q.repo.List(ctx, offset, pageSize)
}
