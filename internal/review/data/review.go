package data

import (
	"context"
	"ecommerce/internal/review/biz"
	"time"

	"gorm.io/gorm"
)

// reviewRepo is the data layer implementation for ReviewRepo.
type reviewRepo struct {
	data *Data
	// log  *log.Helper
}

// toBiz converts a data.Review model to a biz.Review entity.
func (r *Review) toBiz() *biz.Review {
	if r == nil {
		return nil
	}
	return &biz.Review{
		ID:        r.ID,
		ProductID: r.ProductID,
		UserID:    r.UserID,
		Rating:    r.Rating,
		Title:     r.Title,
		Content:   r.Content,
		CreatedAt: r.CreatedAt,
		UpdatedAt: r.UpdatedAt,
	}
}

// fromBiz converts a biz.Review entity to a data.Review model.
func fromBiz(b *biz.Review) *Review {
	if b == nil {
		return nil
	}
	return &Review{
		ProductID: b.ProductID,
		UserID:    b.UserID,
		Rating:    b.Rating,
		Title:     b.Title,
		Content:   b.Content,
	}
}

// CreateReview saves a new review to the database.
func (r *reviewRepo) CreateReview(ctx context.Context, b *biz.Review) (*biz.Review, error) {
	review := fromBiz(b)
	if err := r.data.db.WithContext(ctx).Create(review).Error; err != nil {
		return nil, err
	}
	return review.toBiz(), nil
}

// ListProductReviews lists reviews for a specific product.
func (r *reviewRepo) ListProductReviews(ctx context.Context, productID string, pageSize, pageToken int32) ([]*biz.Review, int32, error) {
	var reviews []Review
	var totalCount int32

	query := r.data.db.WithContext(ctx).Where("product_id = ?", productID)

	// Get total count
	query.Model(&Review{}).Count(int64(&totalCount))

	// Apply pagination
	if pageSize > 0 {
		query = query.Limit(int(pageSize)).Offset(int(pageToken * pageSize))
	}

	if err := query.Find(&reviews).Error; err != nil {
		return nil, 0, err
	}

	bizReviews := make([]*biz.Review, len(reviews))
	for i, rv := range reviews {
		bizReviews[i] = rv.toBiz()
	}

	return bizReviews, totalCount, nil
}

// ListUserReviews lists reviews by a specific user.
func (r *reviewRepo) ListUserReviews(ctx context.Context, userID string, pageSize, pageToken int32) ([]*biz.Review, int32, error) {
	var reviews []Review
	var totalCount int32

	query := r.data.db.WithContext(ctx).Where("user_id = ?", userID)

	// Get total count
	query.Model(&Review{}).Count(int64(&totalCount))

	// Apply pagination
	if pageSize > 0 {
		query = query.Limit(int(pageSize)).Offset(int(pageToken * pageSize))
	}

	if err := query.Find(&reviews).Error; err != nil {
		return nil, 0, err
	}

	bizReviews := make([]*biz.Review, len(reviews))
	for i, rv := range reviews {
		bizReviews[i] = rv.toBiz()
	}

	return bizReviews, totalCount, nil
}

// GetProductRating gets the average rating and total count for a product.
func (r *reviewRepo) GetProductRating(ctx context.Context, productID string) (float64, int32, error) {
	var avgRating float64
	var totalReviews int32

	// Calculate average rating
	if err := r.data.db.WithContext(ctx).Model(&Review{}).Where("product_id = ?", productID).Select("AVG(rating)").Row().Scan(&avgRating); err != nil {
		return 0, 0, err
	}

	// Get total count
	if err := r.data.db.WithContext(ctx).Model(&Review{}).Where("product_id = ?", productID).Count(int64(&totalReviews)).Error; err != nil {
		return 0, 0, err
	}

	return avgRating, totalReviews, nil
}
