package application

import (
	"context"
	"github.com/wyfcoding/ecommerce/internal/review/domain/entity"
	"github.com/wyfcoding/ecommerce/internal/review/domain/repository"
	"errors"

	"log/slog"
)

type ReviewService struct {
	repo   repository.ReviewRepository
	logger *slog.Logger
}

func NewReviewService(repo repository.ReviewRepository, logger *slog.Logger) *ReviewService {
	return &ReviewService{
		repo:   repo,
		logger: logger,
	}
}

// CreateReview 创建评论
func (s *ReviewService) CreateReview(ctx context.Context, userID, productID, orderID, skuID uint64, rating int, content string, images []string) (*entity.Review, error) {
	if rating < 1 || rating > 5 {
		return nil, errors.New("rating must be between 1 and 5")
	}
	if content == "" {
		return nil, errors.New("content cannot be empty")
	}

	review := &entity.Review{
		UserID:    userID,
		ProductID: productID,
		OrderID:   orderID,
		SkuID:     skuID,
		Rating:    rating,
		Content:   content,
		Images:    images,
		Status:    entity.ReviewStatusPending,
	}

	if err := s.repo.Save(ctx, review); err != nil {
		return nil, err
	}

	return review, nil
}

// ApproveReview 审核通过
func (s *ReviewService) ApproveReview(ctx context.Context, id uint64) error {
	review, err := s.repo.Get(ctx, id)
	if err != nil {
		return err
	}
	if review == nil {
		return errors.New("review not found")
	}

	review.Status = entity.ReviewStatusApproved
	return s.repo.Save(ctx, review)
}

// RejectReview 审核拒绝
func (s *ReviewService) RejectReview(ctx context.Context, id uint64) error {
	review, err := s.repo.Get(ctx, id)
	if err != nil {
		return err
	}
	if review == nil {
		return errors.New("review not found")
	}

	review.Status = entity.ReviewStatusRejected
	return s.repo.Save(ctx, review)
}

// ListReviews 评论列表
func (s *ReviewService) ListReviews(ctx context.Context, productID uint64, status *int, page, pageSize int) ([]*entity.Review, int64, error) {
	offset := (page - 1) * pageSize
	var st *entity.ReviewStatus
	if status != nil {
		s := entity.ReviewStatus(*status)
		st = &s
	}
	return s.repo.List(ctx, productID, st, offset, pageSize)
}

// GetProductStats 获取商品评分统计
func (s *ReviewService) GetProductStats(ctx context.Context, productID uint64) (*entity.ProductRatingStats, error) {
	return s.repo.GetProductStats(ctx, productID)
}
