package application

import (
	"context"
	"log/slog"

	"github.com/wyfcoding/ecommerce/internal/logistics/domain"
)

// LogisticsQueryService 负责物流的查询操作。
type LogisticsQueryService struct {
	repo       domain.LogisticsRepository
	readRepo   domain.LogisticsReadRepository
	searchRepo domain.LogisticsSearchRepository
	logger     *slog.Logger
}

// NewLogisticsQueryService 构造函数。
func NewLogisticsQueryService(repo domain.LogisticsRepository, readRepo domain.LogisticsReadRepository, searchRepo domain.LogisticsSearchRepository, logger *slog.Logger) *LogisticsQueryService {
	return &LogisticsQueryService{
		repo:       repo,
		readRepo:   readRepo,
		searchRepo: searchRepo,
		logger:     logger,
	}
}

func (q *LogisticsQueryService) GetLogistics(ctx context.Context, id uint64) (*domain.Logistics, error) {
	if q.readRepo != nil {
		if logistics, err := q.readRepo.GetByID(ctx, id); err == nil && logistics != nil {
			return logistics, nil
		}
	}
	return q.repo.GetByID(ctx, id)
}

func (q *LogisticsQueryService) GetLogisticsByTrackingNo(ctx context.Context, trackingNo string) (*domain.Logistics, error) {
	if q.readRepo != nil {
		if logistics, err := q.readRepo.GetByTrackingNo(ctx, trackingNo); err == nil && logistics != nil {
			return logistics, nil
		}
	}
	if q.searchRepo != nil {
		if logistics, err := q.searchRepo.FindByTrackingNo(ctx, trackingNo); err == nil && logistics != nil {
			return logistics, nil
		}
	}
	return q.repo.GetByTrackingNo(ctx, trackingNo)
}

func (q *LogisticsQueryService) GetLogisticsByOrderID(ctx context.Context, orderID uint64) (*domain.Logistics, error) {
	if q.readRepo != nil {
		if logistics, err := q.readRepo.GetByOrderID(ctx, orderID); err == nil && logistics != nil {
			return logistics, nil
		}
	}
	return q.repo.GetByOrderID(ctx, orderID)
}

func (q *LogisticsQueryService) ListLogistics(ctx context.Context, page, pageSize int) ([]*domain.Logistics, int64, error) {
	offset := (page - 1) * pageSize
	if q.searchRepo != nil {
		list, total, err := q.searchRepo.Search(ctx, nil, nil, nil, nil, offset, pageSize, nil, nil, "created_at")
		if err == nil {
			return list, total, nil
		}
		if q.logger != nil {
			q.logger.WarnContext(ctx, "logistics search fallback to mysql", "error", err)
		}
	}
	return q.repo.List(ctx, offset, pageSize)
}
