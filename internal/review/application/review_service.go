package application

import (
	"context"

	"github.com/wyfcoding/ecommerce/internal/review/domain"
)

// ReviewService 结构体定义了评论平衡相关的应用服务（外观模式）。
// 它将业务逻辑委托给 ReviewManager 和 ReviewQuery 处理。
type ReviewService struct {
	manager *ReviewManager
	query   *ReviewQuery
}

// NewReviewService 创建并返回一个新的 ReviewService 实例。
func NewReviewService(manager *ReviewManager, query *ReviewQuery) *ReviewService {
	return &ReviewService{
		manager: manager,
		query:   query,
	}
}

// CreateReview 提交一条新的评论。
func (s *ReviewService) CreateReview(ctx context.Context, userID, productID, orderID, skuID uint64, rating int, content string, images []string) (*domain.Review, error) {
	return s.manager.CreateReview(ctx, userID, productID, orderID, skuID, rating, content, images)
}

// AuditReview 审核评论。
func (s *ReviewService) AuditReview(ctx context.Context, reviewID uint64, approved bool) error {
	return s.manager.AuditReview(ctx, reviewID, approved)
}

// DeleteReview 删除评论。
func (s *ReviewService) DeleteReview(ctx context.Context, reviewID uint64, userID uint64) error {
	return s.manager.DeleteReview(ctx, reviewID, userID)
}

// GetReview 获取评论详情。
func (s *ReviewService) GetReview(ctx context.Context, id uint64) (*domain.Review, error) {
	return s.query.GetReview(ctx, id)
}

// ListReviews 获取指定商品的评论列表。
func (s *ReviewService) ListReviews(ctx context.Context, productID uint64, status *int, page, pageSize int) ([]*domain.Review, int64, error) {
	return s.query.ListReviews(ctx, productID, status, page, pageSize)
}

// ListUserReviews 获取指定用户的评论列表。
func (s *ReviewService) ListUserReviews(ctx context.Context, userID uint64, page, pageSize int) ([]*domain.Review, int64, error) {
	return s.query.ListUserReviews(ctx, userID, page, pageSize)
}

// GetProductStats 获取商品的评分统计。
func (s *ReviewService) GetProductStats(ctx context.Context, productID uint64) (*domain.ProductRatingStats, error) {
	return s.query.GetProductStats(ctx, productID)
}

// 如果需要兼容旧方法，则保留，但门面应优先使用更清晰的名称。
func (s *ReviewService) ApproveReview(ctx context.Context, id uint64) error {
	return s.manager.AuditReview(ctx, id, true)
}

func (s *ReviewService) RejectReview(ctx context.Context, id uint64) error {
	return s.manager.AuditReview(ctx, id, false)
}
