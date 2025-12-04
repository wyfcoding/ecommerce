package persistence

import (
	"context"
	"errors" // 导入标准错误处理库。

	"github.com/wyfcoding/ecommerce/internal/review/domain/entity"     // 导入评论领域的实体定义。
	"github.com/wyfcoding/ecommerce/internal/review/domain/repository" // 导入评论领域的仓储接口。

	"gorm.io/gorm" // 导入GORM ORM框架。
)

type reviewRepository struct {
	db *gorm.DB // GORM数据库连接实例。
}

// NewReviewRepository 创建并返回一个新的 reviewRepository 实例。
func NewReviewRepository(db *gorm.DB) repository.ReviewRepository {
	return &reviewRepository{db: db}
}

// Save 将评论实体保存到数据库。
// 如果实体已存在，则更新；如果不存在，则创建。
func (r *reviewRepository) Save(ctx context.Context, review *entity.Review) error {
	return r.db.WithContext(ctx).Save(review).Error
}

// Get 根据ID从数据库获取评论记录。
// 如果记录未找到，则返回nil。
func (r *reviewRepository) Get(ctx context.Context, id uint64) (*entity.Review, error) {
	var review entity.Review
	if err := r.db.WithContext(ctx).First(&review, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // 如果记录未找到，返回nil。
		}
		return nil, err // 其他错误则返回。
	}
	return &review, nil
}

// List 从数据库列出指定商品ID的所有评论记录，支持通过状态过滤和分页。
func (r *reviewRepository) List(ctx context.Context, productID uint64, status *entity.ReviewStatus, offset, limit int) ([]*entity.Review, int64, error) {
	var list []*entity.Review
	var total int64

	db := r.db.WithContext(ctx).Model(&entity.Review{})
	if productID > 0 { // 如果提供了商品ID，则按商品ID过滤。
		db = db.Where("product_id = ?", productID)
	}
	if status != nil { // 如果提供了状态，则按状态过滤。
		db = db.Where("status = ?", *status)
	}

	// 统计总记录数。
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 应用分页和排序。
	if err := db.Offset(offset).Limit(limit).Order("created_at desc").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// Delete 根据ID从数据库删除评论记录。
// GORM默认进行软删除。
func (r *reviewRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&entity.Review{}, id).Error
}

// GetProductStats 计算并获取指定商品的评分统计数据。
func (r *reviewRepository) GetProductStats(ctx context.Context, productID uint64) (*entity.ProductRatingStats, error) {
	var stats entity.ProductRatingStats
	stats.ProductID = productID

	// 通过SQL聚合查询计算每个评分等级的评论数量。
	// 注意：这是一个简化的实现。在生产环境中，为了性能，这些统计数据可能需要：
	// 1. 缓存起来。
	// 2. 使用物化视图或定期批处理计算。
	rows, err := r.db.WithContext(ctx).Model(&entity.Review{}).
		Select("rating, count(*) as count").                                            // 选择评分和计数。
		Where("product_id = ? AND status = ?", productID, entity.ReviewStatusApproved). // 只统计已通过审核的评论。
		Group("rating").                                                                // 按评分分组。
		Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var totalRating int64 // 用于计算总评分。
	for rows.Next() {
		var rating, count int
		if err := rows.Scan(&rating, &count); err != nil {
			return nil, err
		}
		stats.TotalReviews += count          // 累加总评论数。
		totalRating += int64(rating * count) // 累加总评分。
		// 根据评分等级更新对应的计数。
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

	// 计算平均评分。
	if stats.TotalReviews > 0 {
		stats.AverageRating = float64(totalRating) / float64(stats.TotalReviews)
	}

	return &stats, nil
}
