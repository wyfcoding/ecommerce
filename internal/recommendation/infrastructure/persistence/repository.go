package persistence

import (
	"context"
	"ecommerce/internal/recommendation/domain/entity"
	"ecommerce/internal/recommendation/domain/repository"
	"errors"

	"gorm.io/gorm"
)

type recommendationRepository struct {
	db *gorm.DB
}

func NewRecommendationRepository(db *gorm.DB) repository.RecommendationRepository {
	return &recommendationRepository{db: db}
}

// 推荐结果
func (r *recommendationRepository) SaveRecommendation(ctx context.Context, rec *entity.Recommendation) error {
	return r.db.WithContext(ctx).Save(rec).Error
}

func (r *recommendationRepository) ListRecommendations(ctx context.Context, userID uint64, recType *entity.RecommendationType, limit int) ([]*entity.Recommendation, error) {
	var list []*entity.Recommendation
	db := r.db.WithContext(ctx).Where("user_id = ?", userID)
	if recType != nil {
		db = db.Where("recommendation_type = ?", *recType)
	}
	if err := db.Order("score desc").Limit(limit).Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *recommendationRepository) DeleteRecommendations(ctx context.Context, userID uint64, recType *entity.RecommendationType) error {
	db := r.db.WithContext(ctx).Where("user_id = ?", userID)
	if recType != nil {
		db = db.Where("recommendation_type = ?", *recType)
	}
	return db.Delete(&entity.Recommendation{}).Error
}

// 用户偏好
func (r *recommendationRepository) SaveUserPreference(ctx context.Context, pref *entity.UserPreference) error {
	return r.db.WithContext(ctx).Save(pref).Error
}

func (r *recommendationRepository) GetUserPreference(ctx context.Context, userID uint64) (*entity.UserPreference, error) {
	var pref entity.UserPreference
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&pref).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &pref, nil
}

// 商品相似度
func (r *recommendationRepository) SaveProductSimilarity(ctx context.Context, sim *entity.ProductSimilarity) error {
	return r.db.WithContext(ctx).Save(sim).Error
}

func (r *recommendationRepository) ListSimilarProducts(ctx context.Context, productID uint64, limit int) ([]*entity.ProductSimilarity, error) {
	var list []*entity.ProductSimilarity
	if err := r.db.WithContext(ctx).Where("product_id = ?", productID).Order("similarity desc").Limit(limit).Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

// 用户行为
func (r *recommendationRepository) SaveUserBehavior(ctx context.Context, behavior *entity.UserBehavior) error {
	return r.db.WithContext(ctx).Save(behavior).Error
}

func (r *recommendationRepository) ListUserBehaviors(ctx context.Context, userID uint64, limit int) ([]*entity.UserBehavior, error) {
	var list []*entity.UserBehavior
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Order("timestamp desc").Limit(limit).Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}
