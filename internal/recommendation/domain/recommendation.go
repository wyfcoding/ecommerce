// 变更说明：完善推荐系统领域模型，增加推荐算法、协同过滤、用户画像等高级功能
package domain

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

// RecommendationType 推荐类型
type RecommendationType string

const (
	RecommendationTypePersonalized RecommendationType = "PERSONALIZED"
	RecommendationTypeHot          RecommendationType = "HOT"
	RecommendationTypeSimilar      RecommendationType = "SIMILAR"
	RecommendationTypeRelated      RecommendationType = "RELATED"
	RecommendationTypeNewArrival   RecommendationType = "NEW_ARRIVAL"
	RecommendationTypeTrending     RecommendationType = "TRENDING"
	RecommendationTypeFlashSale    RecommendationType = "FLASH_SALE"
	RecommendationTypeCrossSell    RecommendationType = "CROSS_SELL"
	RecommendationTypeUpSell       RecommendationType = "UP_SELL"
)

// AlgorithmType 推荐算法类型
type AlgorithmType string

const (
	AlgorithmCollaborativeFiltering AlgorithmType = "COLLABORATIVE_FILTERING"
	AlgorithmContentBased            AlgorithmType = "CONTENT_BASED"
	AlgorithmHybrid                  AlgorithmType = "HYBRID"
	AlgorithmDeepLearning            AlgorithmType = "DEEP_LEARNING"
	AlgorithmRuleBased               AlgorithmType = "RULE_BASED"
	AlgorithmPopularity              AlgorithmType = "POPULARITY"
)

// Recommendation 推荐结果
type Recommendation struct {
	ID               uint64             `json:"id"`
	UserID           uint64             `json:"user_id"`
	ProductID        uint64             `json:"product_id"`
	RecommendationType RecommendationType `json:"recommendation_type"`
	Algorithm        AlgorithmType      `json:"algorithm"`
	Score            float64            `json:"score"`
	Confidence       float64            `json:"confidence"`
	Reason           string             `json:"reason"`
	ReasonCode       string             `json:"reason_code"`
	Position         int                `json:"position"`
	Context          string             `json:"context"`
	ExpiresAt        *time.Time         `json:"expires_at"`
	Clicked          bool               `json:"clicked"`
	ClickedAt        *time.Time         `json:"clicked_at"`
	Converted        bool               `json:"converted"`
	ConvertedAt      *time.Time         `json:"converted_at"`
	CreatedAt        time.Time          `json:"created_at"`
	UpdatedAt        time.Time          `json:"updated_at"`
}

// UserPreference 用户偏好
type UserPreference struct {
	ID            uint64      `json:"id"`
	UserID        uint64      `json:"user_id"`
	CategoryIDs   []uint64    `json:"category_ids"`
	BrandIDs      []uint64    `json:"brand_ids"`
	PriceRangeMin uint64      `json:"price_range_min"`
	PriceRangeMax uint64      `json:"price_range_max"`
	Tags          StringArray `json:"tags"`
	Keywords      StringArray `json:"keywords"`
	Weight        float64     `json:"weight"`
	Source        string      `json:"source"`
	ExpiresAt     *time.Time  `json:"expires_at"`
	CreatedAt     time.Time   `json:"created_at"`
	UpdatedAt     time.Time   `json:"updated_at"`
}

// UserBehavior 用户行为
type UserBehavior struct {
	ID          uint64    `json:"id"`
	UserID      uint64    `json:"user_id"`
	ProductID   uint64    `json:"product_id"`
	CategoryID  uint64    `json:"category_id"`
	BrandID     uint64    `json:"brand_id"`
	Action      string    `json:"action"`
	Weight      float64   `json:"weight"`
	Duration    int64     `json:"duration"`
	Quantity    int       `json:"quantity"`
	Amount      float64   `json:"amount"`
	SessionID   string    `json:"session_id"`
	DeviceID    string    `json:"device_id"`
	Platform    string    `json:"platform"`
	IPAddress   string    `json:"ip_address"`
	UserAgent   string    `json:"user_agent"`
	Referrer    string    `json:"referrer"`
	Timestamp   time.Time `json:"timestamp"`
	CreatedAt   time.Time `json:"created_at"`
}

// ProductSimilarity 商品相似度
type ProductSimilarity struct {
	ID               uint64    `json:"id"`
	ProductID        uint64    `json:"product_id"`
	SimilarProductID uint64    `json:"similar_product_id"`
	Similarity       float64   `json:"similarity"`
	Algorithm        AlgorithmType `json:"algorithm"`
	Dimensions       StringArray `json:"dimensions"`
	UpdatedAt        time.Time `json:"updated_at"`
	CreatedAt        time.Time `json:"created_at"`
}

// UserProfile 用户画像
type UserProfile struct {
	ID                 uint64    `json:"id"`
	UserID             uint64    `json:"user_id"`
	AgeGroup           string    `json:"age_group"`
	Gender             string    `json:"gender"`
	IncomeLevel        string    `json:"income_level"`
	ShoppingFrequency  string    `json:"shopping_frequency"`
	PriceSensitivity   float64   `json:"price_sensitivity"`
	BrandLoyalty       float64   `json:"brand_loyalty"`
	CategoryPreference StringArray `json:"category_preference"`
	BrandPreference    StringArray `json:"brand_preference"`
	ShoppingHabits     StringArray `json:"shopping_habits"`
	DevicePreference   string    `json:"device_preference"`
	ActiveHours        StringArray `json:"active_hours"`
	LTVScore           float64   `json:"ltv_score"`
	ChurnRisk          float64   `json:"churn_risk"`
	EngagementScore    float64   `json:"engagement_score"`
	LastUpdated        time.Time `json:"last_updated"`
	CreatedAt          time.Time `json:"created_at"`
}

// ProductFeature 商品特征
type ProductFeature struct {
	ID               uint64      `json:"id"`
	ProductID        uint64      `json:"product_id"`
	CategoryID       uint64      `json:"category_id"`
	BrandID          uint64      `json:"brand_id"`
	Price            float64     `json:"price"`
	PriceLevel       int         `json:"price_level"`
	SalesCount       int64       `json:"sales_count"`
	ViewCount        int64       `json:"view_count"`
	CartCount        int64       `json:"cart_count"`
	ConversionRate   float64     `json:"conversion_rate"`
	Rating           float64     `json:"rating"`
	ReviewCount      int64       `json:"review_count"`
	ReturnRate       float64     `json:"return_rate"`
	Tags             StringArray `json:"tags"`
	Keywords         StringArray `json:"keywords"`
	Attributes       string      `json:"attributes"`
	EmbeddingVector  []float64   `json:"embedding_vector"`
	PopularityScore  float64     `json:"popularity_score"`
	QualityScore     float64     `json:"quality_score"`
	LastUpdated      time.Time   `json:"last_updated"`
	CreatedAt        time.Time   `json:"created_at"`
}

// RecommendationRequest 推荐请求
type RecommendationRequest struct {
	UserID             uint64            `json:"user_id"`
	ProductID          uint64            `json:"product_id"`
	CategoryID         uint64            `json:"category_id"`
	RecommendationType RecommendationType `json:"recommendation_type"`
	Limit              int               `json:"limit"`
	Offset             int               `json:"offset"`
	Context            string            `json:"context"`
	ExcludeProductIDs  []uint64          `json:"exclude_product_ids"`
	Filters            map[string]string `json:"filters"`
}

// RecommendationResult 推荐结果
type RecommendationResult struct {
	Items        []*Recommendation `json:"items"`
	Total        int64             `json:"total"`
	Algorithm    AlgorithmType     `json:"algorithm"`
	GeneratedAt  time.Time         `json:"generated_at"`
	CacheHit     bool              `json:"cache_hit"`
	ProcessTime  int64             `json:"process_time_ms"`
}

// CollaborativeFilteringModel 协同过滤模型
type CollaborativeFilteringModel struct {
	UserItemMatrix     map[uint64]map[uint64]float64 `json:"user_item_matrix"`
	ItemUserMatrix     map[uint64]map[uint64]float64 `json:"item_user_matrix"`
	UserSimilarity     map[uint64]map[uint64]float64 `json:"user_similarity"`
	ItemSimilarity     map[uint64]map[uint64]float64 `json:"item_similarity"`
	UserMeans          map[uint64]float64            `json:"user_means"`
	ItemMeans          map[uint64]float64            `json:"item_means"`
	GlobalMean         float64                       `json:"global_mean"`
	LastUpdated        time.Time                     `json:"last_updated"`
}

// RecommendationConfig 推荐配置
type RecommendationConfig struct {
	ID                        uint64    `json:"id"`
	Name                      string    `json:"name"`
	RecommendationType        RecommendationType `json:"recommendation_type"`
	PrimaryAlgorithm          AlgorithmType `json:"primary_algorithm"`
	FallbackAlgorithm         AlgorithmType `json:"fallback_algorithm"`
	MinScoreThreshold         float64   `json:"min_score_threshold"`
	MaxItems                  int       `json:"max_items"`
	DiversityFactor           float64   `json:"diversity_factor"`
	NoveltyFactor             float64   `json:"novelty_factor"`
	FreshnessWeight           float64   `json:"freshness_weight"`
	PopularityWeight          float64   `json:"popularity_weight"`
	PersonalizationWeight     float64   `json:"personalization_weight"`
	CacheTTLSeconds           int       `json:"cache_ttl_seconds"`
	EnableABTest              bool      `json:"enable_ab_test"`
	ABTestRatio               float64   `json:"ab_test_ratio"`
	Status                    string    `json:"status"`
	CreatedAt                 time.Time `json:"created_at"`
	UpdatedAt                 time.Time `json:"updated_at"`
}

// ABTestExperiment AB测试实验
type ABTestExperiment struct {
	ID           uint64    `json:"id"`
	Name         string    `json:"name"`
	Description  string    `json:"description"`
	ControlGroup string    `json:"control_group"`
	TestGroup    string    `json:"test_group"`
	TrafficRatio float64   `json:"traffic_ratio"`
	StartTime    time.Time `json:"start_time"`
	EndTime      time.Time `json:"end_time"`
	Status       string    `json:"status"`
	Metrics      string    `json:"metrics"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// StringArray 字符串数组类型
type StringArray []string

func (a StringArray) Value() (driver.Value, error) {
	if a == nil {
		return nil, nil
	}
	return json.Marshal(a)
}

func (a *StringArray) Scan(value any) error {
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

// NewUserBehavior 创建用户行为
func NewUserBehavior(userID, productID uint64, action string, weight float64) *UserBehavior {
	now := time.Now()
	return &UserBehavior{
		UserID:    userID,
		ProductID: productID,
		Action:    action,
		Weight:    weight,
		Timestamp: now,
		CreatedAt: now,
	}
}

// CalculateWeight 根据行为类型计算权重
func CalculateBehaviorWeight(action string) float64 {
	weights := map[string]float64{
		"view":     1.0,
		"click":    2.0,
		"cart":     3.0,
		"wishlist": 2.5,
		"buy":      5.0,
		"review":   4.0,
		"share":    3.5,
		"favorite": 2.5,
	}
	
	if w, ok := weights[action]; ok {
		return w
	}
	return 1.0
}

// NewProductSimilarity 创建商品相似度
func NewProductSimilarity(productID, similarProductID uint64, similarity float64, algorithm AlgorithmType) *ProductSimilarity {
	now := time.Now()
	return &ProductSimilarity{
		ProductID:        productID,
		SimilarProductID: similarProductID,
		Similarity:       similarity,
		Algorithm:        algorithm,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
}

// NewRecommendation 创建推荐
func NewRecommendation(userID, productID uint64, recType RecommendationType, algorithm AlgorithmType, score float64, reason string) *Recommendation {
	now := time.Now()
	return &Recommendation{
		UserID:            userID,
		ProductID:         productID,
		RecommendationType: recType,
		Algorithm:         algorithm,
		Score:             score,
		Reason:            reason,
		CreatedAt:         now,
		UpdatedAt:         now,
	}
}

// MarkClicked 标记点击
func (r *Recommendation) MarkClicked() {
	r.Clicked = true
	now := time.Now()
	r.ClickedAt = &now
	r.UpdatedAt = now
}

// MarkConverted 标记转化
func (r *Recommendation) MarkConverted() {
	r.Converted = true
	now := time.Time{}
	r.ConvertedAt = &now
	r.UpdatedAt = now
}

// IsExpired 检查是否过期
func (r *Recommendation) IsExpired() bool {
	if r.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*r.ExpiresAt)
}

// CalculateCosineSimilarity 计算余弦相似度
func CalculateCosineSimilarity(vec1, vec2 map[uint64]float64) float64 {
	dotProduct := 0.0
	norm1 := 0.0
	norm2 := 0.0
	
	for key, val1 := range vec1 {
		if val2, exists := vec2[key]; exists {
			dotProduct += val1 * val2
		}
		norm1 += val1 * val1
	}
	
	for _, val := range vec2 {
		norm2 += val * val
	}
	
	if norm1 == 0 || norm2 == 0 {
		return 0
	}
	
	return dotProduct / (sqrt(norm1) * sqrt(norm2))
}

// CalculatePearsonCorrelation 计算皮尔逊相关系数
func CalculatePearsonCorrelation(vec1, vec2 map[uint64]float64, mean1, mean2 float64) float64 {
	numerator := 0.0
	denom1 := 0.0
	denom2 := 0.0
	
	for key, val1 := range vec1 {
		if val2, exists := vec2[key]; exists {
			diff1 := val1 - mean1
			diff2 := val2 - mean2
			numerator += diff1 * diff2
			denom1 += diff1 * diff1
			denom2 += diff2 * diff2
		}
	}
	
	if denom1 == 0 || denom2 == 0 {
		return 0
	}
	
	return numerator / (sqrt(denom1) * sqrt(denom2))
}

// 辅助函数
func sqrt(x float64) float64 {
	if x < 0 {
		return 0
	}
	z := x
	for i := 0; i < 10; i++ {
		z = (z + x/z) / 2
	}
	return z
}

// 错误定义
var (
	ErrUserNotFound         = errors.New("user not found")
	ErrProductNotFound      = errors.New("product not found")
	ErrInsufficientData     = errors.New("insufficient data for recommendation")
	ErrAlgorithmNotReady    = errors.New("recommendation algorithm not ready")
	ErrInvalidRequest       = errors.New("invalid recommendation request")
)

// GenerateReasonCode 生成推荐理由代码
func GenerateReasonCode(recType RecommendationType, algorithm AlgorithmType) string {
	return fmt.Sprintf("%s_%s", recType, algorithm)
}
