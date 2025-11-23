package entity

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"gorm.io/gorm"
)

// RecommendationType 推荐类型
type RecommendationType string

const (
	RecommendationTypePersonalized RecommendationType = "PERSONALIZED" // 个性化推荐
	RecommendationTypeHot          RecommendationType = "HOT"          // 热门推荐
	RecommendationTypeSimilar      RecommendationType = "SIMILAR"      // 相似推荐
	RecommendationTypeRelated      RecommendationType = "RELATED"      // 关联推荐
)

// Recommendation 推荐实体
type Recommendation struct {
	gorm.Model
	UserID             uint64             `gorm:"not null;index;comment:用户ID" json:"user_id"`
	RecommendationType RecommendationType `gorm:"type:varchar(32);not null;comment:推荐类型" json:"recommendation_type"`
	ProductID          uint64             `gorm:"not null;index;comment:商品ID" json:"product_id"`
	Score              float64            `gorm:"type:decimal(10,4);not null;comment:推荐分数" json:"score"`
	Reason             string             `gorm:"type:varchar(255);comment:推荐理由" json:"reason"`
}

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

// UserPreference 用户偏好实体
type UserPreference struct {
	gorm.Model
	UserID     uint64      `gorm:"uniqueIndex;not null;comment:用户ID" json:"user_id"`
	CategoryID uint64      `gorm:"index;comment:偏好类目ID" json:"category_id"`
	BrandID    uint64      `gorm:"index;comment:偏好品牌ID" json:"brand_id"`
	PriceMin   uint64      `gorm:"comment:价格区间下限(分)" json:"price_min"`
	PriceMax   uint64      `gorm:"comment:价格区间上限(分)" json:"price_max"`
	Tags       StringArray `gorm:"type:json;comment:偏好标签" json:"tags"`
	Weight     float64     `gorm:"type:decimal(10,4);not null;default:1.0;comment:权重" json:"weight"`
}

// ProductSimilarity 商品相似度实体
type ProductSimilarity struct {
	gorm.Model
	ProductID        uint64  `gorm:"uniqueIndex:idx_product_similar;not null;comment:商品ID" json:"product_id"`
	SimilarProductID uint64  `gorm:"uniqueIndex:idx_product_similar;not null;comment:相似商品ID" json:"similar_product_id"`
	Similarity       float64 `gorm:"type:decimal(10,4);not null;comment:相似度" json:"similarity"`
}

// UserBehavior 用户行为记录 (用于推荐计算)
type UserBehavior struct {
	gorm.Model
	UserID    uint64    `gorm:"index;not null;comment:用户ID" json:"user_id"`
	ProductID uint64    `gorm:"index;not null;comment:商品ID" json:"product_id"`
	Action    string    `gorm:"type:varchar(32);not null;comment:行为类型(view,click,cart,buy)" json:"action"`
	Weight    float64   `gorm:"type:decimal(10,4);not null;default:1.0;comment:权重" json:"weight"`
	Timestamp time.Time `gorm:"not null;comment:发生时间" json:"timestamp"`
}
