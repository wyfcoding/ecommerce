package mysql

import (
	"time"

	"github.com/wyfcoding/ecommerce/internal/recommendation/domain"
	"gorm.io/gorm"
)

// RecommendationModel 推荐写模型（持久化专用）。
type RecommendationModel struct {
	gorm.Model
	UserID             uint64                    `gorm:"not null;index;comment:用户ID"`
	RecommendationType domain.RecommendationType `gorm:"type:varchar(32);not null;comment:推荐类型"`
	ProductID          uint64                    `gorm:"not null;index;comment:商品ID"`
	Score              float64                   `gorm:"type:decimal(10,4);not null;comment:推荐分数"`
	Reason             string                    `gorm:"type:varchar(255);comment:推荐理由"`
}

func (RecommendationModel) TableName() string {
	return "recommendations"
}

// UserPreferenceModel 用户偏好写模型（持久化专用）。
type UserPreferenceModel struct {
	gorm.Model
	UserID     uint64             `gorm:"uniqueIndex;not null;comment:用户ID"`
	CategoryID uint64             `gorm:"index;comment:偏好类目ID"`
	BrandID    uint64             `gorm:"index;comment:偏好品牌ID"`
	PriceMin   uint64             `gorm:"comment:价格区间下限(分)"`
	PriceMax   uint64             `gorm:"comment:价格区间上限(分)"`
	Tags       domain.StringArray `gorm:"type:json;comment:偏好标签"`
	Weight     float64            `gorm:"type:decimal(10,4);not null;default:1.0;comment:权重"`
}

func (UserPreferenceModel) TableName() string {
	return "user_preferences"
}

// ProductSimilarityModel 商品相似度写模型（持久化专用）。
type ProductSimilarityModel struct {
	gorm.Model
	ProductID        uint64  `gorm:"uniqueIndex:idx_product_similar;not null;comment:商品ID"`
	SimilarProductID uint64  `gorm:"uniqueIndex:idx_product_similar;not null;comment:相似商品ID"`
	Similarity       float64 `gorm:"type:decimal(10,4);not null;comment:相似度"`
}

func (ProductSimilarityModel) TableName() string {
	return "product_similarities"
}

// UserBehaviorModel 用户行为写模型（持久化专用）。
type UserBehaviorModel struct {
	gorm.Model
	UserID    uint64    `gorm:"index;not null;comment:用户ID"`
	ProductID uint64    `gorm:"index;not null;comment:商品ID"`
	Action    string    `gorm:"type:varchar(32);not null;comment:行为类型(view,click,cart,buy)"`
	Weight    float64   `gorm:"type:decimal(10,4);not null;default:1.0;comment:权重"`
	Timestamp time.Time `gorm:"not null;comment:发生时间"`
}

func (UserBehaviorModel) TableName() string {
	return "user_behaviors"
}

func toRecommendationModel(rec *domain.Recommendation) *RecommendationModel {
	if rec == nil {
		return nil
	}
	return &RecommendationModel{
		Model: gorm.Model{
			ID:        uint(rec.ID),
			CreatedAt: rec.CreatedAt,
			UpdatedAt: rec.UpdatedAt,
		},
		UserID:             rec.UserID,
		RecommendationType: rec.RecommendationType,
		ProductID:          rec.ProductID,
		Score:              rec.Score,
		Reason:             rec.Reason,
	}
}

func toRecommendation(model *RecommendationModel) *domain.Recommendation {
	if model == nil {
		return nil
	}
	return &domain.Recommendation{
		ID:                 uint64(model.ID),
		CreatedAt:          model.CreatedAt,
		UpdatedAt:          model.UpdatedAt,
		UserID:             model.UserID,
		RecommendationType: model.RecommendationType,
		ProductID:          model.ProductID,
		Score:              model.Score,
		Reason:             model.Reason,
	}
}

func toUserPreferenceModel(pref *domain.UserPreference) *UserPreferenceModel {
	if pref == nil {
		return nil
	}
	var categoryID uint64
	if len(pref.CategoryIDs) > 0 {
		categoryID = pref.CategoryIDs[0]
	}
	var brandID uint64
	if len(pref.BrandIDs) > 0 {
		brandID = pref.BrandIDs[0]
	}
	return &UserPreferenceModel{
		Model: gorm.Model{
			ID:        uint(pref.ID),
			CreatedAt: pref.CreatedAt,
			UpdatedAt: pref.UpdatedAt,
		},
		UserID:     pref.UserID,
		CategoryID: categoryID,
		BrandID:    brandID,
		PriceMin:   pref.PriceRangeMin,
		PriceMax:   pref.PriceRangeMax,
		Tags:       pref.Tags,
		Weight:     pref.Weight,
	}
}

func toUserPreference(model *UserPreferenceModel) *domain.UserPreference {
	if model == nil {
		return nil
	}
	pref := &domain.UserPreference{
		ID:            uint64(model.ID),
		CreatedAt:     model.CreatedAt,
		UpdatedAt:     model.UpdatedAt,
		UserID:        model.UserID,
		PriceRangeMin: model.PriceMin,
		PriceRangeMax: model.PriceMax,
		Tags:          model.Tags,
		Weight:        model.Weight,
	}
	if model.CategoryID != 0 {
		pref.CategoryIDs = []uint64{model.CategoryID}
	}
	if model.BrandID != 0 {
		pref.BrandIDs = []uint64{model.BrandID}
	}
	return pref
}

func toProductSimilarityModel(sim *domain.ProductSimilarity) *ProductSimilarityModel {
	if sim == nil {
		return nil
	}
	return &ProductSimilarityModel{
		Model: gorm.Model{
			ID:        uint(sim.ID),
			CreatedAt: sim.CreatedAt,
			UpdatedAt: sim.UpdatedAt,
		},
		ProductID:        sim.ProductID,
		SimilarProductID: sim.SimilarProductID,
		Similarity:       sim.Similarity,
	}
}

func toProductSimilarity(model *ProductSimilarityModel) *domain.ProductSimilarity {
	if model == nil {
		return nil
	}
	return &domain.ProductSimilarity{
		ID:               uint64(model.ID),
		CreatedAt:        model.CreatedAt,
		UpdatedAt:        model.UpdatedAt,
		ProductID:        model.ProductID,
		SimilarProductID: model.SimilarProductID,
		Similarity:       model.Similarity,
	}
}

func toUserBehaviorModel(behavior *domain.UserBehavior) *UserBehaviorModel {
	if behavior == nil {
		return nil
	}
	return &UserBehaviorModel{
		Model: gorm.Model{
			ID:        uint(behavior.ID),
			CreatedAt: behavior.CreatedAt,
		},
		UserID:    behavior.UserID,
		ProductID: behavior.ProductID,
		Action:    behavior.Action,
		Weight:    behavior.Weight,
		Timestamp: behavior.Timestamp,
	}
}

func toUserBehavior(model *UserBehaviorModel) *domain.UserBehavior {
	if model == nil {
		return nil
	}
	return &domain.UserBehavior{
		ID:        uint64(model.ID),
		CreatedAt: model.CreatedAt,
		UserID:    model.UserID,
		ProductID: model.ProductID,
		Action:    model.Action,
		Weight:    model.Weight,
		Timestamp: model.Timestamp,
	}
}
