package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"ecommerce/internal/review/model"
)

// ReviewRepository 定义了评论数据仓库的接口
type ReviewRepository interface {
	CreateReview(ctx context.Context, review *model.Review) error
	GetReview(ctx context.Context, id uint) (*model.Review, error)
	ListReviewsByProduct(ctx context.Context, productID uint, page, pageSize int) ([]model.Review, int64, error)
	UpdateReviewStatus(ctx context.Context, id uint, status model.ReviewStatus) error
	CreateComment(ctx context.Context, comment *model.Comment) error

	// UpdateProductReviewStats 是一个关键的原子操作
	UpdateProductReviewStats(ctx context.Context, productID uint, ratingChange, reviewCountChange int) error
}

// reviewRepository 是接口的具体实现
type reviewRepository struct {
	db *gorm.DB
}

// NewReviewRepository 创建一个新的 reviewRepository 实例
func NewReviewRepository(db *gorm.DB) ReviewRepository {
	return &reviewRepository{db: db}
}

func (r *reviewRepository) CreateReview(ctx context.Context, review *model.Review) error {
	if err := r.db.WithContext(ctx).Create(review).Error; err != nil {
		return fmt.Errorf("数据库创建评论失败: %w", err)
	}
	return nil
}

func (r *reviewRepository) GetReview(ctx context.Context, id uint) (*model.Review, error) {
	var review model.Review
	if err := r.db.WithContext(ctx).Preload("Comments").First(&review, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("数据库查询评论失败: %w", err)
	}
	return &review, nil
}

func (r *reviewRepository) ListReviewsByProduct(ctx context.Context, productID uint, page, pageSize int) ([]model.Review, int64, error) {
	var reviews []model.Review
	var total int64

	db := r.db.WithContext(ctx).Model(&model.Review{}).Where("product_id = ? AND status = ?", productID, model.StatusApproved)

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("数据库统计评论数量失败: %w", err)
	}

	offset := (page - 1) * pageSize
	if err := db.Order("created_at desc").Offset(offset).Limit(pageSize).Find(&reviews).Error; err != nil {
		return nil, 0, fmt.Errorf("数据库列出评论失败: %w", err)
	}

	return reviews, total, nil
}

func (r *reviewRepository) UpdateReviewStatus(ctx context.Context, id uint, status model.ReviewStatus) error {
	result := r.db.WithContext(ctx).Model(&model.Review{}).Where("id = ?", id).Update("status", status)
	if result.Error != nil {
		return fmt.Errorf("数据库更新评论状态失败: %w", result.Error)
	}
	return nil
}

func (r *reviewRepository) CreateComment(ctx context.Context, comment *model.Comment) error {
	if err := r.db.WithContext(ctx).Create(comment).Error; err != nil {
		return fmt.Errorf("数据库创建回复失败: %w", err)
	}
	return nil
}

// UpdateProductReviewStats 原子地更新产品的评论统计信息
func (r *reviewRepository) UpdateProductReviewStats(ctx context.Context, productID uint, ratingChange, reviewCountChange int) error {
	// 使用事务来保证原子性
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var stats model.ProductReviewStats

		// 1. 锁定行以进行更新
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&stats, productID).Error; err != nil {
			// 如果记录不存在，则创建它
			if err == gorm.ErrRecordNotFound {
				stats = model.ProductReviewStats{ProductID: productID}
			} else {
				return err
			}
		}

		// 2. 计算新的总评论数和总分数
		newTotalReviews := stats.TotalReviews + reviewCountChange
		// (旧的平均分 * 旧的总数) + 新增的分数
		newTotalScore := (stats.AverageRating * float64(stats.TotalReviews)) + float64(ratingChange)

		// 3. 计算新的平均分
		var newAverageRating float64
		if newTotalReviews > 0 {
			newAverageRating = newTotalScore / float64(newTotalReviews)
		} else {
			newAverageRating = 0
		}

		stats.TotalReviews = newTotalReviews
		stats.AverageRating = newAverageRating

		// 4. TODO: 更新评分分布图 (rating_counts)

		// 5. 保存或创建统计记录
		return tx.Save(&stats).Error
	})
}
