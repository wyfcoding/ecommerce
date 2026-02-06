package application

import (
	"context"
	"errors"
	"log/slog"

	"github.com/wyfcoding/ecommerce/internal/logistics/domain"
)

// LogisticsProjectionService 负责将事件转换为读模型更新。
type LogisticsProjectionService struct {
	repo       domain.LogisticsRepository
	readRepo   domain.LogisticsReadRepository
	searchRepo domain.LogisticsSearchRepository
	logger     *slog.Logger
}

// NewLogisticsProjectionService 创建物流投影服务。
func NewLogisticsProjectionService(repo domain.LogisticsRepository, readRepo domain.LogisticsReadRepository, searchRepo domain.LogisticsSearchRepository, logger *slog.Logger) *LogisticsProjectionService {
	return &LogisticsProjectionService{
		repo:       repo,
		readRepo:   readRepo,
		searchRepo: searchRepo,
		logger:     logger,
	}
}

// OnLogisticsCreated 处理物流创建事件。
func (s *LogisticsProjectionService) OnLogisticsCreated(ctx context.Context, event *domain.LogisticsCreatedEvent) error {
	return s.refreshLogistics(ctx, uint64(event.LogisticsID))
}

// OnStatusUpdated 处理物流状态更新事件。
func (s *LogisticsProjectionService) OnStatusUpdated(ctx context.Context, event *domain.LogisticsStatusUpdatedEvent) error {
	return s.refreshLogistics(ctx, uint64(event.LogisticsID))
}

// OnTraceAdded 处理物流轨迹新增事件。
func (s *LogisticsProjectionService) OnTraceAdded(ctx context.Context, event *domain.LogisticsTraceAddedEvent) error {
	return s.refreshLogistics(ctx, uint64(event.LogisticsID))
}

// OnRiderAssigned 处理骑手指派事件。
func (s *LogisticsProjectionService) OnRiderAssigned(ctx context.Context, event *domain.RiderAssignedEvent) error {
	return s.refreshLogistics(ctx, uint64(event.LogisticsID))
}

func (s *LogisticsProjectionService) refreshLogistics(ctx context.Context, id uint64) error {
	logistics, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrLogisticsNotFound) {
			return s.cleanupReadModel(ctx, id, "", 0)
		}
		s.logger.ErrorContext(ctx, "failed to load logistics for projection", "logistics_id", id, "error", err)
		return err
	}

	if logistics == nil {
		return s.cleanupReadModel(ctx, id, "", 0)
	}

	if s.readRepo != nil {
		if err := s.readRepo.Save(ctx, logistics); err != nil {
			s.logger.ErrorContext(ctx, "failed to save logistics read model", "logistics_id", id, "error", err)
			return err
		}
	}
	if s.searchRepo != nil {
		if err := s.searchRepo.Index(ctx, logistics); err != nil {
			s.logger.ErrorContext(ctx, "failed to index logistics search model", "logistics_id", id, "error", err)
			return err
		}
	}
	return nil
}

func (s *LogisticsProjectionService) cleanupReadModel(ctx context.Context, id uint64, trackingNo string, orderID uint64) error {
	if s.readRepo != nil {
		// 尝试读取补齐 trackingNo/orderID 以清理多 key
		if trackingNo == "" || orderID == 0 {
			if cached, err := s.readRepo.GetByID(ctx, id); err == nil && cached != nil {
				if trackingNo == "" {
					trackingNo = cached.TrackingNo
				}
				if orderID == 0 {
					orderID = cached.OrderID
				}
			}
		}
		_ = s.readRepo.Delete(ctx, id, trackingNo, orderID)
	}
	if s.searchRepo != nil {
		_ = s.searchRepo.Delete(ctx, id)
	}
	return nil
}
