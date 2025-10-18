package repository

import (
	"context"

	"ecommerce/internal/review/model"
)

// ReviewRepo defines the data storage interface for reviews.
// The business layer depends on this interface, not on a concrete data implementation.
type ReviewRepo interface {
	CreateReview(ctx context.Context, review *model.Review) (*model.Review, error)
	ListProductReviews(ctx context.Context, productID string, pageSize, pageToken int32) ([]*model.Review, int32, error)
	ListUserReviews(ctx context.Context, userID string, pageSize, pageToken int32) ([]*model.Review, int32, error)
	GetProductRating(ctx context.Context, productID string) (float64, int32, error)
}