// 生成摘要：新增评论读模型投影服务，消费事件后刷新 Redis/ES 读侧。
// 假设：读模型以 review_id 为主键，写模型为最终一致性来源。
package application

import (
	"context"
	"log/slog"

	"github.com/wyfcoding/ecommerce/internal/review/domain"
)

// ReviewProjectionService 负责将事件转换为读模型更新。
type ReviewProjectionService struct {
	repo       domain.ReviewRepository
	readRepo   domain.ReviewReadRepository
	searchRepo domain.ReviewSearchRepository
	logger     *slog.Logger
}

// NewReviewProjectionService 创建评论投影服务。
func NewReviewProjectionService(repo domain.ReviewRepository, readRepo domain.ReviewReadRepository, searchRepo domain.ReviewSearchRepository, logger *slog.Logger) *ReviewProjectionService {
	return &ReviewProjectionService{
		repo:       repo,
		readRepo:   readRepo,
		searchRepo: searchRepo,
		logger:     logger,
	}
}

// OnReviewCreated 处理评论创建事件。
func (s *ReviewProjectionService) OnReviewCreated(ctx context.Context, event *domain.ReviewCreatedEvent) error {
	return s.refreshReview(ctx, event.ReviewID)
}

// OnReviewUpdated 处理评论更新事件。
func (s *ReviewProjectionService) OnReviewUpdated(ctx context.Context, event *domain.ReviewUpdatedEvent) error {
	return s.refreshReview(ctx, event.ReviewID)
}

// OnReviewDeleted 处理评论删除事件。
func (s *ReviewProjectionService) OnReviewDeleted(ctx context.Context, event *domain.ReviewDeletedEvent) error {
	if event == nil {
		return nil
	}
	if s.readRepo != nil {
		_ = s.readRepo.Delete(ctx, event.ReviewID)
	}
	if s.searchRepo != nil {
		_ = s.searchRepo.Delete(ctx, event.ReviewID)
	}
	return nil
}

func (s *ReviewProjectionService) refreshReview(ctx context.Context, reviewID uint64) error {
	if reviewID == 0 {
		return nil
	}
	review, err := s.repo.Get(ctx, reviewID)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to load review for projection", "review_id", reviewID, "error", err)
		return err
	}
	if review == nil {
		if s.readRepo != nil {
			_ = s.readRepo.Delete(ctx, reviewID)
		}
		if s.searchRepo != nil {
			_ = s.searchRepo.Delete(ctx, reviewID)
		}
		return nil
	}

	if s.readRepo != nil {
		if err := s.readRepo.Save(ctx, review); err != nil {
			s.logger.ErrorContext(ctx, "failed to save review read model", "review_id", reviewID, "error", err)
			return err
		}
	}
	if s.searchRepo != nil {
		if err := s.searchRepo.Index(ctx, review); err != nil {
			s.logger.ErrorContext(ctx, "failed to index review", "review_id", reviewID, "error", err)
			return err
		}
	}
	return nil
}
