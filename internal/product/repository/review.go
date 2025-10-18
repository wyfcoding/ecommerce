package repository

import (
	"context"

	"ecommerce/internal/product/model"
)

// ReviewRepo is a Review repo.
type ReviewRepo interface {
	CreateReview(ctx context.Context, review *model.Review) (*model.Review, error)
	ListReviews(ctx context.Context, spuID uint64, pageSize, pageNum uint32, minRating *uint32) ([]*model.Review, uint64, error)
	DeleteReview(ctx context.Context, id uint64, userID uint64) error
}
