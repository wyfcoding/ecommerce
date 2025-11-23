package entity

import (
	"database/sql/driver"
	"encoding/json"
	"errors"

	"gorm.io/gorm"
)

// ReviewStatus 评论状态
type ReviewStatus int

const (
	ReviewStatusPending  ReviewStatus = 1
	ReviewStatusApproved ReviewStatus = 2
	ReviewStatusRejected ReviewStatus = 3
)

// StringArray defines a slice of strings that implements sql.Scanner and driver.Valuer
type StringArray []string

func (a StringArray) Value() (driver.Value, error) {
	if a == nil {
		return nil, nil
	}
	return json.Marshal(a)
}

func (a *StringArray) Scan(value interface{}) error {
	if value == nil {
		*a = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, a)
}

// Review 评论聚合根
type Review struct {
	gorm.Model
	UserID    uint64       `gorm:"not null;index;comment:用户ID" json:"user_id"`
	ProductID uint64       `gorm:"not null;index;comment:商品ID" json:"product_id"`
	OrderID   uint64       `gorm:"not null;index;comment:订单ID" json:"order_id"`
	SkuID     uint64       `gorm:"not null;index;comment:SKU ID" json:"sku_id"`
	Rating    int          `gorm:"not null;comment:评分(1-5)" json:"rating"`
	Content   string       `gorm:"type:text;not null;comment:评论内容" json:"content"`
	Images    StringArray  `gorm:"type:json;comment:图片列表" json:"images"`
	Status    ReviewStatus `gorm:"type:tinyint;not null;default:1;comment:状态" json:"status"`
	LikeCount int          `gorm:"not null;default:0;comment:点赞数" json:"like_count"`
}

// ProductRatingStats 商品评分统计值对象
type ProductRatingStats struct {
	ProductID     uint64  `json:"product_id"`
	AverageRating float64 `json:"average_rating"`
	TotalReviews  int     `json:"total_reviews"`
	Rating5Count  int     `json:"rating_5_count"`
	Rating4Count  int     `json:"rating_4_count"`
	Rating3Count  int     `json:"rating_3_count"`
	Rating2Count  int     `json:"rating_2_count"`
	Rating1Count  int     `json:"rating_1_count"`
}
