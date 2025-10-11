package data

import (
	"time"

	"gorm.io/gorm"
)

// ProductRecommendation represents a record of a smart product selection recommendation.
type ProductRecommendation struct {
	gorm.Model
	MerchantID      string    `gorm:"index;not null;comment:商家ID" json:"merchantId"`
	ProductID       uint64    `gorm:"index;not null;comment:商品ID" json:"productId"`
	ProductName     string    `gorm:"size:255;comment:商品名称" json:"productName"`
	Score           float64   `gorm:"not null;comment:推荐分数" json:"score"`
	Reason          string    `gorm:"type:text;comment:推荐理由" json:"reason"`
	ContextFeatures string    `gorm:"type:json;comment:选品上下文特征 (JSON字符串)" json:"contextFeatures"`
	RecommendedAt   time.Time `gorm:"not null;comment:推荐时间" json:"recommendedAt"`
}

// TableName specifies the table name for ProductRecommendation.
func (ProductRecommendation) TableName() string {
	return "smart_product_recommendations"
}
