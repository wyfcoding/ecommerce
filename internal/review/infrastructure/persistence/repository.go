package persistence

import (
	"context"
	"github.com/wyfcoding/ecommerce/internal/review/domain/entity"
	"github.com/wyfcoding/ecommerce/internal/review/domain/repository"
	"errors"

	"gorm.io/gorm"
)

type reviewRepository struct {
	db *gorm.DB
}

func NewReviewRepository(db *gorm.DB) repository.ReviewRepository {
	return &reviewRepository{db: db}
}

func (r *reviewRepository) Save(ctx context.Context, review *entity.Review) error {
	return r.db.WithContext(ctx).Save(review).Error
}

func (r *reviewRepository) Get(ctx context.Context, id uint64) (*entity.Review, error) {
	var review entity.Review
	if err := r.db.WithContext(ctx).First(&review, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &review, nil
}

func (r *reviewRepository) List(ctx context.Context, productID uint64, status *entity.ReviewStatus, offset, limit int) ([]*entity.Review, int64, error) {
	var list []*entity.Review
	var total int64

	db := r.db.WithContext(ctx).Model(&entity.Review{})
	if productID > 0 {
		db = db.Where("product_id = ?", productID)
	}
	if status != nil {
		db = db.Where("status = ?", *status)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := db.Offset(offset).Limit(limit).Order("created_at desc").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

func (r *reviewRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&entity.Review{}, id).Error
}

func (r *reviewRepository) GetProductStats(ctx context.Context, productID uint64) (*entity.ProductRatingStats, error) {
	var stats entity.ProductRatingStats
	stats.ProductID = productID

	// Calculate stats using aggregation
	// This is a simplified implementation. In production, you might want to cache this or use a materialized view.
	rows, err := r.db.WithContext(ctx).Model(&entity.Review{}).
		Select("rating, count(*) as count").
		Where("product_id = ? AND status = ?", productID, entity.ReviewStatusApproved).
		Group("rating").
		Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var totalRating int64
	for rows.Next() {
		var rating, count int
		if err := rows.Scan(&rating, &count); err != nil {
			return nil, err
		}
		stats.TotalReviews += count
		totalRating += int64(rating * count)
		switch rating {
		case 5:
			stats.Rating5Count = count
		case 4:
			stats.Rating4Count = count
		case 3:
			stats.Rating3Count = count
		case 2:
			stats.Rating2Count = count
		case 1:
			stats.Rating1Count = count
		}
	}

	if stats.TotalReviews > 0 {
		stats.AverageRating = float64(totalRating) / float64(stats.TotalReviews)
	}

	return &stats, nil
}
