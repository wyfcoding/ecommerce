package mysql

import (
	"context"
	"errors"
	"log/slog"

	"github.com/wyfcoding/ecommerce/internal/review/domain"
	"gorm.io/gorm"
)

type reviewRepository struct {
	db     *gorm.DB
	logger *slog.Logger
}

// NewReviewRepository 创建并返回一个新的 reviewRepository 实例。
func NewReviewRepository(db *gorm.DB, logger *slog.Logger) domain.ReviewRepository {
	return &reviewRepository{
		db:     db,
		logger: logger,
	}
}

func (r *reviewRepository) BeginTx(ctx context.Context) any {
	return r.db.WithContext(ctx).Begin()
}

func (r *reviewRepository) CommitTx(tx any) error {
	return tx.(*gorm.DB).Commit().Error
}

func (r *reviewRepository) RollbackTx(tx any) error {
	return tx.(*gorm.DB).Rollback().Error
}

func (r *reviewRepository) WithTx(ctx context.Context, fn func(tx any) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(tx)
	})
}

// Save 将评论实体保存到数据库。
func (r *reviewRepository) Save(ctx context.Context, review *domain.Review) error {
	model := toReviewModel(review)
	if model == nil {
		return nil
	}
	if err := r.db.WithContext(ctx).Save(model).Error; err != nil {
		return err
	}
	*review = *toReview(model)
	return nil
}

// SaveInTx 在事务中保存评论。
func (r *reviewRepository) SaveInTx(ctx context.Context, tx any, review *domain.Review) error {
	model := toReviewModel(review)
	if model == nil {
		return nil
	}
	if err := tx.(*gorm.DB).WithContext(ctx).Save(model).Error; err != nil {
		return err
	}
	*review = *toReview(model)
	return nil
}

// Get 根据ID从数据库获取评论记录。
func (r *reviewRepository) Get(ctx context.Context, id uint64) (*domain.Review, error) {
	var model ReviewModel
	if err := r.db.WithContext(ctx).First(&model, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toReview(&model), nil
}

// List 从数据库列出指定商品ID的所有评论记录。
func (r *reviewRepository) List(ctx context.Context, productID uint64, status *domain.ReviewStatus, offset, limit int) ([]*domain.Review, int64, error) {
	var list []*ReviewModel
	var total int64

	db := r.db.WithContext(ctx).Model(&ReviewModel{})
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

	results := make([]*domain.Review, 0, len(list))
	for _, model := range list {
		results = append(results, toReview(model))
	}
	return results, total, nil
}

// ListByUser 从数据库列出指定用户ID的所有评论记录。
func (r *reviewRepository) ListByUser(ctx context.Context, userID uint64, offset, limit int) ([]*domain.Review, int64, error) {
	var list []*ReviewModel
	var total int64

	db := r.db.WithContext(ctx).Model(&ReviewModel{}).Where("user_id = ?", userID)

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := db.Offset(offset).Limit(limit).Order("created_at desc").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	results := make([]*domain.Review, 0, len(list))
	for _, model := range list {
		results = append(results, toReview(model))
	}
	return results, total, nil
}

// Delete 根据ID从数据库删除评论记录。
func (r *reviewRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&ReviewModel{}, id).Error
}

// DeleteInTx 在事务中删除评论记录。
func (r *reviewRepository) DeleteInTx(ctx context.Context, tx any, id uint64) error {
	return tx.(*gorm.DB).WithContext(ctx).Delete(&ReviewModel{}, id).Error
}

// GetProductStats 计算并获取指定商品的评分统计数据。
func (r *reviewRepository) GetProductStats(ctx context.Context, productID uint64) (*domain.ProductRatingStats, error) {
	var stats domain.ProductRatingStats
	stats.ProductID = productID

	rows, err := r.db.WithContext(ctx).Model(&ReviewModel{}).
		Select("rating, count(*) as count").
		Where("product_id = ? AND status = ?", productID, domain.ReviewStatusApproved).
		Group("rating").
		Rows()
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			r.logger.Error("failed to close rows", "error", closeErr)
		}
	}()

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
