package application

import (
	"context"
	"log/slog"

	"github.com/wyfcoding/ecommerce/internal/review/domain"
)

// ReviewQueryService 处理评论模块的查询操作。
type ReviewQueryService struct {
	repo       domain.ReviewRepository
	readRepo   domain.ReviewReadRepository
	searchRepo domain.ReviewSearchRepository
	logger     *slog.Logger
}

// NewReviewQueryService 创建并返回一个新的 ReviewQueryService 实例。
func NewReviewQueryService(repo domain.ReviewRepository, readRepo domain.ReviewReadRepository, searchRepo domain.ReviewSearchRepository, logger *slog.Logger) *ReviewQueryService {
	return &ReviewQueryService{repo: repo, readRepo: readRepo, searchRepo: searchRepo, logger: logger}
}

// GetReview 根据ID获取评论详情。
func (q *ReviewQueryService) GetReview(ctx context.Context, id uint64) (*domain.Review, error) {
	if q.readRepo != nil {
		if review, err := q.readRepo.GetByID(ctx, id); err == nil && review != nil {
			return review, nil
		}
	}
	return q.repo.Get(ctx, id)
}

// ListReviews 获取指定商品的评论列表。
func (q *ReviewQueryService) ListReviews(ctx context.Context, productID uint64, status *int, page, pageSize int) ([]*domain.Review, int64, error) {
	offset := (page - 1) * pageSize
	var reviewStatus *domain.ReviewStatus
	if status != nil {
		s := domain.ReviewStatus(*status)
		reviewStatus = &s
	}
	if q.searchRepo != nil {
		var productPtr *uint64
		if productID > 0 {
			productPtr = &productID
		}
		return q.searchRepo.Search(ctx, productPtr, nil, reviewStatus, offset, pageSize, "")
	}
	return q.repo.List(ctx, productID, reviewStatus, offset, pageSize)
}

// ListUserReviews 获取指定用户的评论列表。
func (q *ReviewQueryService) ListUserReviews(ctx context.Context, userID uint64, page, pageSize int) ([]*domain.Review, int64, error) {
	offset := (page - 1) * pageSize
	if q.searchRepo != nil {
		userPtr := &userID
		return q.searchRepo.Search(ctx, nil, userPtr, nil, offset, pageSize, "")
	}
	return q.repo.ListByUser(ctx, userID, offset, pageSize)
}

// GetProductStats 获取商品的评分统计。
func (q *ReviewQueryService) GetProductStats(ctx context.Context, productID uint64) (*domain.ProductRatingStats, error) {
	if q.readRepo != nil {
		stats, err := q.readRepo.GetProductStats(ctx, productID)
		if err == nil && stats != nil {
			return stats, nil
		}
	}
	return q.repo.GetProductStats(ctx, productID)
}
