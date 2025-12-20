package application

import (
	"context"
	"log/slog"

	"github.com/wyfcoding/ecommerce/internal/logistics/domain"
)

// LogisticsQuery 处理物流的读操作。
type LogisticsQuery struct {
	repo   domain.LogisticsRepository
	logger *slog.Logger
}

// NewLogisticsQuery 负责处理 NewLogistics 相关的读操作和查询逻辑。
func NewLogisticsQuery(repo domain.LogisticsRepository, logger *slog.Logger) *LogisticsQuery {
	return &LogisticsQuery{
		repo:   repo,
		logger: logger,
	}
}

// GetLogistics 获取指定ID的物流信息。
func (q *LogisticsQuery) GetLogistics(ctx context.Context, id uint64) (*domain.Logistics, error) {
	return q.repo.GetByID(ctx, id)
}

// GetLogisticsByTrackingNo 根据运单号获取物流信息。
func (q *LogisticsQuery) GetLogisticsByTrackingNo(ctx context.Context, trackingNo string) (*domain.Logistics, error) {
	return q.repo.GetByTrackingNo(ctx, trackingNo)
}

// ListLogistics 获取物流单列表。
func (q *LogisticsQuery) ListLogistics(ctx context.Context, page, pageSize int) ([]*domain.Logistics, int64, error) {
	offset := (page - 1) * pageSize
	return q.repo.List(ctx, offset, pageSize)
}
