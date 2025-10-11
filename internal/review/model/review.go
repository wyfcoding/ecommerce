package biz

import (
	"context"
	"errors"
	"time"
)

// ErrReviewNotFound is a specific error for when a review is not found.
var ErrReviewNotFound = errors.New("review not found")

// Review represents a review entity in the business layer.
type Review struct {
	ID        uint
	ProductID string
	UserID    string
	Rating    int32 // 1-5 stars
	Title     string
	Content   string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// ReviewRepo defines the data storage interface for reviews.
// The business layer depends on this interface, not on a concrete data implementation.
type ReviewRepo interface {
	CreateReview(ctx context.Context, review *Review) (*Review, error)
	ListProductReviews(ctx context.Context, productID string, pageSize, pageToken int32) ([]*Review, int32, error)
	ListUserReviews(ctx context.Context, userID string, pageSize, pageToken int32) ([]*Review, int32, error)
	GetProductRating(ctx context.Context, productID string) (float64, int32, error)
}

// ReviewUsecase is the use case for review-related operations.
// It orchestrates the business logic.
type ReviewUsecase struct {
	repo ReviewRepo
	// You can also inject other dependencies like a logger
}

// NewReviewUsecase creates a new ReviewUsecase.
func NewReviewUsecase(repo ReviewRepo) *ReviewUsecase {
	return &ReviewUsecase{repo: repo}
}

// CreateReview creates a new review.
func (uc *ReviewUsecase) CreateReview(ctx context.Context, productID, userID string, rating int32, title, content string) (*Review, error) {
	// Here you can add business logic before creating, e.g., validation, spam checks, etc.
	review := &Review{
		ProductID: productID,
		UserID:    userID,
		Rating:    rating,
		Title:     title,
		Content:   content,
	}
	return uc.repo.CreateReview(ctx, review)
}

// ListProductReviews lists reviews for a specific product.
func (uc *ReviewUsecase) ListProductReviews(ctx context.Context, productID string, pageSize, pageToken int32) ([]*Review, int32, error) {
	return uc.repo.ListProductReviews(ctx, productID, pageSize, pageToken)
}

// ListUserReviews lists reviews by a specific user.
func (uc *ReviewUsecase) ListUserReviews(ctx context.Context, userID string, pageSize, pageToken int32) ([]*Review, int32, error) {
	return uc.repo.ListUserReviews(ctx, userID, pageSize, pageToken)
}

// GetProductRating gets the average rating and total count for a product.
func (uc *ReviewUsecase) GetProductRating(ctx context.Context, productID string) (float64, int32, error) {
	return uc.repo.GetProductRating(ctx, productID)
}
