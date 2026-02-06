// 生成摘要：新增推荐读模型投影服务，消费事件后刷新 Redis/ES 读侧。
// 假设：读模型以 user_id + type 为主键，写模型为最终一致性来源。
package application

import (
	"context"
	"log/slog"

	"github.com/wyfcoding/ecommerce/internal/recommendation/domain"
)

// RecommendationProjectionService 负责将事件转换为读模型更新。
type RecommendationProjectionService struct {
	repo       domain.RecommendationRepository
	readRepo   domain.RecommendationReadRepository
	searchRepo domain.RecommendationSearchRepository
	logger     *slog.Logger
}

// NewRecommendationProjectionService 创建推荐投影服务。
func NewRecommendationProjectionService(repo domain.RecommendationRepository, readRepo domain.RecommendationReadRepository, searchRepo domain.RecommendationSearchRepository, logger *slog.Logger) *RecommendationProjectionService {
	return &RecommendationProjectionService{
		repo:       repo,
		readRepo:   readRepo,
		searchRepo: searchRepo,
		logger:     logger,
	}
}

// OnRecommendationChanged 处理推荐变化事件。
func (s *RecommendationProjectionService) OnRecommendationChanged(ctx context.Context, event *domain.RecommendationChangedEvent) error {
	return s.refreshRecommendations(ctx, event.UserID, &event.RecommendationType)
}

// OnRecommendationDeleted 处理推荐删除事件。
func (s *RecommendationProjectionService) OnRecommendationDeleted(ctx context.Context, event *domain.RecommendationDeletedEvent) error {
	if s.readRepo != nil {
		_ = s.readRepo.DeleteRecommendations(ctx, event.UserID, &event.RecommendationType)
	}
	if s.searchRepo != nil {
		_ = s.searchRepo.DeleteByUserAndType(ctx, event.UserID, &event.RecommendationType)
	}
	return nil
}

func (s *RecommendationProjectionService) refreshRecommendations(ctx context.Context, userID uint64, recType *domain.RecommendationType) error {
	if userID == 0 {
		return nil
	}

	recs, err := s.repo.ListRecommendations(ctx, userID, recType, 1000)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to load recommendations for projection", "user_id", userID, "error", err)
		return err
	}

	if s.readRepo != nil {
		if err := s.readRepo.SaveRecommendations(ctx, userID, recType, recs); err != nil {
			s.logger.ErrorContext(ctx, "failed to save recommendation read model", "user_id", userID, "error", err)
			return err
		}
	}
	if s.searchRepo != nil {
		if err := s.searchRepo.DeleteByUserAndType(ctx, userID, recType); err != nil {
			s.logger.WarnContext(ctx, "failed to clear recommendation search model", "user_id", userID, "error", err)
		}
		for _, rec := range recs {
			if err := s.searchRepo.Index(ctx, rec); err != nil {
				s.logger.ErrorContext(ctx, "failed to index recommendation", "user_id", userID, "error", err)
				return err
			}
		}
	}

	return nil
}
