package mysql

import (
	"github.com/wyfcoding/ecommerce/internal/review/domain"
	"gorm.io/gorm"
)

// ReviewModel 评论写模型（持久化专用）。
type ReviewModel struct {
	gorm.Model
	UserID    uint64              `gorm:"not null;index;comment:用户ID"`
	ProductID uint64              `gorm:"not null;index;comment:商品ID"`
	OrderID   uint64              `gorm:"not null;index;comment:订单ID"`
	SkuID     uint64              `gorm:"not null;index;comment:SKU ID"`
	Rating    int                 `gorm:"not null;comment:评分(1-5)"`
	Content   string              `gorm:"type:text;not null;comment:评论内容"`
	Images    domain.StringArray  `gorm:"type:json;comment:图片列表"`
	Status    domain.ReviewStatus `gorm:"type:tinyint;not null;default:1;comment:状态"`
	LikeCount int                 `gorm:"not null;default:0;comment:点赞数"`
}

func (ReviewModel) TableName() string {
	return "reviews"
}

func toReviewModel(review *domain.Review) *ReviewModel {
	if review == nil {
		return nil
	}
	return &ReviewModel{
		Model: gorm.Model{
			ID:        uint(review.ID),
			CreatedAt: review.CreatedAt,
			UpdatedAt: review.UpdatedAt,
		},
		UserID:    review.UserID,
		ProductID: review.ProductID,
		OrderID:   review.OrderID,
		SkuID:     review.SkuID,
		Rating:    review.Rating,
		Content:   review.Content,
		Images:    review.Images,
		Status:    review.Status,
		LikeCount: review.LikeCount,
	}
}

func toReview(model *ReviewModel) *domain.Review {
	if model == nil {
		return nil
	}
	return &domain.Review{
		ID:        uint64(model.ID),
		CreatedAt: model.CreatedAt,
		UpdatedAt: model.UpdatedAt,
		UserID:    model.UserID,
		ProductID: model.ProductID,
		OrderID:   model.OrderID,
		SkuID:     model.SkuID,
		Rating:    model.Rating,
		Content:   model.Content,
		Images:    model.Images,
		Status:    model.Status,
		LikeCount: model.LikeCount,
	}
}
