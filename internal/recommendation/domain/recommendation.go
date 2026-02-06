package domain

import (
	"database/sql/driver" // 导入数据库驱动接口。
	"encoding/json"       // 导入JSON编码/解码库。
	"errors"              // 导入标准错误处理库。
	"time"                // 导入时间包。
)

// RecommendationType 定义了推荐的类型。
type RecommendationType string

const (
	RecommendationTypePersonalized RecommendationType = "PERSONALIZED" // 个性化推荐：根据用户的行为和偏好生成。
	RecommendationTypeHot          RecommendationType = "HOT"          // 热门推荐：基于商品整体流行度。
	RecommendationTypeSimilar      RecommendationType = "SIMILAR"      // 相似推荐：与用户当前查看或已购商品相似。
	RecommendationTypeRelated      RecommendationType = "RELATED"      // 关联推荐：通常与用户购买的其他商品一起购买。
)

// Recommendation 实体代表一个推荐结果。
// 它包含了推荐给哪个用户、推荐类型、推荐的商品、推荐分数和理由。
type Recommendation struct {
	ID                 uint64             `json:"id"`
	CreatedAt          time.Time          `json:"created_at"`
	UpdatedAt          time.Time          `json:"updated_at"`
	UserID             uint64             `json:"user_id"`             // 接收推荐的用户ID。
	RecommendationType RecommendationType `json:"recommendation_type"` // 推荐类型。
	ProductID          uint64             `json:"product_id"`          // 推荐的商品ID。
	Score              float64            `json:"score"`               // 推荐的得分，用于排序。
	Reason             string             `json:"reason"`              // 推荐给用户的理由。
}

// StringArray 定义了一个字符串切片类型，实现了 sql.Scanner 和 driver.Valuer 接口，
// 允许GORM将Go的 []string 类型作为JSON字符串存储到数据库，并从数据库读取。
type StringArray []string

// Value 实现 driver.Valuer 接口，将 StringArray 转换为数据库可以存储的值（JSON字节数组）。
func (a StringArray) Value() (driver.Value, error) {
	if a == nil {
		return nil, nil
	}
	return json.Marshal(a) // 将切片编码为JSON字节数组。
}

// Scan 实现 sql.Scanner 接口，从数据库读取值并转换为 StringArray。
func (a *StringArray) Scan(value any) error {
	if value == nil {
		*a = nil
		return nil
	}
	bytes, ok := value.([]byte) // 期望数据库返回字节数组。
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, a) // 将JSON字节数组解码为切片。
}

// UserPreference 实体代表用户的个性化偏好设置。
// 它可以用于指导个性化推荐的生成。
type UserPreference struct {
	ID         uint64      `json:"id"`
	CreatedAt  time.Time   `json:"created_at"`
	UpdatedAt  time.Time   `json:"updated_at"`
	UserID     uint64      `json:"user_id"`     // 用户ID。
	CategoryID uint64      `json:"category_id"` // 用户偏好的商品类目ID。
	BrandID    uint64      `json:"brand_id"`    // 用户偏好的商品品牌ID。
	PriceMin   uint64      `json:"price_min"`   // 用户偏好的价格区间下限。
	PriceMax   uint64      `json:"price_max"`   // 用户偏好的价格区间上限。
	Tags       StringArray `json:"tags"`        // 用户偏好的商品标签列表，存储为JSON。
	Weight     float64     `json:"weight"`      // 用户偏好的权重，用于影响推荐算法。
}

// ProductSimilarity 实体记录了商品之间的相似度。
// 用于实现相似商品推荐功能。
type ProductSimilarity struct {
	ID               uint64    `json:"id"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
	ProductID        uint64    `json:"product_id"`         // 商品ID。
	SimilarProductID uint64    `json:"similar_product_id"` // 相似商品ID。
	Similarity       float64   `json:"similarity"`         // 两个商品之间的相似度分数。
}

// UserBehavior 实体记录了用户的行为数据。
// 这些数据是推荐系统生成推荐的基石。
type UserBehavior struct {
	ID        uint64    `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	UserID    uint64    `json:"user_id"`    // 用户ID。
	ProductID uint64    `json:"product_id"` // 发生行为的商品ID。
	Action    string    `json:"action"`     // 行为类型，例如“view”“click”“cart”“buy”。
	Weight    float64   `json:"weight"`     // 行为权重，用于推荐算法。
	Timestamp time.Time `json:"timestamp"`  // 行为发生的时间。
}
