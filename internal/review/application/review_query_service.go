package application

import (
	"context"
	"log/slog"

	"github.com/wyfcoding/ecommerce/internal/review/domain"
)

// ReviewQueryService 处理评论模块的查询操作。
type ReviewQueryService struct {
	repo   domain.ReviewRepository
	logger *slog.Logger
}

// NewReviewQueryService 创建并返回一个新的 ReviewQueryService 实例。
func NewReviewQueryService(repo domain.ReviewRepository, logger *slog.Logger) *ReviewQueryService {
	return &ReviewQueryService{repo: repo, logger: logger}
}

// GetReview 根据ID获取评论详情。
func (q *ReviewQueryService) GetReview(ctx context.Context, id uint64) (*domain.Review, error) {
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
	return q.repo.List(ctx, productID, reviewStatus, offset, pageSize)
}

// ListUserReviews 获取指定用户的评论列表。
func (q *ReviewQueryService) ListUserReviews(ctx context.Context, userID uint64, page, pageSize int) ([]*domain.Review, int64, error) {
	offset := (page - 1) * pageSize
	return q.repo.ListByUser(ctx, userID, offset, pageSize)
}

// GetProductStats 获取商品的评分统计。
func (q *ReviewQueryService) GetProductStats(ctx context.Context, productID uint64) (*domain.ProductRatingStats, error) {
	return q.repo.GetProductStats(ctx, productID)
}
