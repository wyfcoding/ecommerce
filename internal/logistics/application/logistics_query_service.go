package application

import (
	"context"

	"github.com/wyfcoding/ecommerce/internal/logistics/domain"
)

// LogisticsQueryService 负责物流的查询操作。
type LogisticsQueryService struct {
	repo domain.LogisticsRepository
}

// NewLogisticsQueryService 构造函数。
func NewLogisticsQueryService(repo domain.LogisticsRepository) *LogisticsQueryService {
	return &LogisticsQueryService{repo: repo}
}

func (q *LogisticsQueryService) GetLogistics(ctx context.Context, id uint64) (*domain.Logistics, error) {
	return q.repo.GetByID(ctx, id)
}

func (q *LogisticsQueryService) GetLogisticsByTrackingNo(ctx context.Context, trackingNo string) (*domain.Logistics, error) {
	return q.repo.GetByTrackingNo(ctx, trackingNo)
}

func (q *LogisticsQueryService) GetLogisticsByOrderID(ctx context.Context, orderID uint64) (*domain.Logistics, error) {
	return q.repo.GetByOrderID(ctx, orderID)
}

func (q *LogisticsQueryService) ListLogistics(ctx context.Context, page, pageSize int) ([]*domain.Logistics, int64, error) {
	offset := (page - 1) * pageSize
	return q.repo.List(ctx, offset, pageSize)
}
