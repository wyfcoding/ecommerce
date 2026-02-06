// 生成摘要：新增AI模型读模型投影服务，消费事件后刷新 Redis/ES 读侧。
package application

import (
	"context"
	"errors"
	"log/slog"

	"github.com/wyfcoding/ecommerce/internal/aimodel/domain"
)

// AIModelProjectionService 负责将AI模型事件投影到读模型。
type AIModelProjectionService struct {
	repo       domain.AIModelRepository
	readRepo   domain.AIModelReadRepository
	searchRepo domain.AIModelSearchRepository
	logger     *slog.Logger
}

// NewAIModelProjectionService 创建AI模型投影服务。
func NewAIModelProjectionService(
	repo domain.AIModelRepository,
	readRepo domain.AIModelReadRepository,
	searchRepo domain.AIModelSearchRepository,
	logger *slog.Logger,
) *AIModelProjectionService {
	return &AIModelProjectionService{
		repo:       repo,
		readRepo:   readRepo,
		searchRepo: searchRepo,
		logger:     logger,
	}
}

func (s *AIModelProjectionService) OnModelCreated(ctx context.Context, event *domain.AIModelCreatedEvent) error {
	if event == nil {
		return nil
	}
	return s.refreshModel(ctx, event.ModelID)
}

func (s *AIModelProjectionService) OnModelStatusUpdated(ctx context.Context, event *domain.AIModelStatusUpdatedEvent) error {
	if event == nil {
		return nil
	}
	return s.refreshModel(ctx, event.ModelID)
}

func (s *AIModelProjectionService) refreshModel(ctx context.Context, modelID uint64) error {
	if s.readRepo == nil && s.searchRepo == nil {
		return nil
	}

	model, err := s.repo.GetModel(ctx, modelID)
	if err != nil {
		if errors.Is(err, domain.ErrModelNotFound) {
			if s.readRepo != nil {
				if cached, _ := s.readRepo.GetByID(ctx, modelID); cached != nil {
					_ = s.readRepo.Delete(ctx, modelID, cached.ModelNo)
				} else {
					_ = s.readRepo.Delete(ctx, modelID, "")
				}
			}
			if s.searchRepo != nil {
				_ = s.searchRepo.Delete(ctx, modelID)
			}
			return nil
		}
		s.logger.ErrorContext(ctx, "failed to load model for projection", "model_id", modelID, "error", err)
		return err
	}

	if s.readRepo != nil {
		if err := s.readRepo.Save(ctx, model); err != nil {
			s.logger.ErrorContext(ctx, "failed to save model read cache", "model_id", modelID, "error", err)
			return err
		}
	}
	if s.searchRepo != nil {
		if err := s.searchRepo.Index(ctx, model); err != nil {
			s.logger.ErrorContext(ctx, "failed to index model", "model_id", modelID, "error", err)
			return err
		}
	}
	return nil
}
