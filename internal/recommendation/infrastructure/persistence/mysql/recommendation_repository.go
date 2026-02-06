package mysql

import (
	"context"
	"errors"

	"github.com/wyfcoding/ecommerce/internal/recommendation/domain"
	"gorm.io/gorm"
)

type recommendationRepository struct {
	db *gorm.DB
}

// NewRecommendationRepository 创建并返回一个新的 RecommendationRepository 实例。
func NewRecommendationRepository(db *gorm.DB) domain.RecommendationRepository {
	return &recommendationRepository{db: db}
}

// BeginTx 开始事务
func (r *recommendationRepository) BeginTx(ctx context.Context) any {
	return r.db.WithContext(ctx).Begin()
}

// CommitTx 提交事务
func (r *recommendationRepository) CommitTx(tx any) error {
	return tx.(*gorm.DB).Commit().Error
}

// RollbackTx 回滚事务
func (r *recommendationRepository) RollbackTx(tx any) error {
	return tx.(*gorm.DB).Rollback().Error
}

// WithTx 事务包装器
func (r *recommendationRepository) WithTx(ctx context.Context, fn func(tx any) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(tx)
	})
}

// --- 推荐结果 (Recommendation methods) ---

func (r *recommendationRepository) SaveRecommendation(ctx context.Context, rec *domain.Recommendation) error {
	model := toRecommendationModel(rec)
	if model == nil {
		return nil
	}
	if err := r.db.WithContext(ctx).Save(model).Error; err != nil {
		return err
	}
	*rec = *toRecommendation(model)
	return nil
}

func (r *recommendationRepository) ListRecommendations(ctx context.Context, userID uint64, recType *domain.RecommendationType, limit int) ([]*domain.Recommendation, error) {
	var list []*RecommendationModel
	db := r.db.WithContext(ctx).Where("user_id = ?", userID)
	if recType != nil {
		db = db.Where("recommendation_type = ?", *recType)
	}
	if limit <= 0 {
		limit = 10
	}
	if err := db.Order("score desc").Limit(limit).Find(&list).Error; err != nil {
		return nil, err
	}
	results := make([]*domain.Recommendation, 0, len(list))
	for _, model := range list {
		results = append(results, toRecommendation(model))
	}
	return results, nil
}

func (r *recommendationRepository) DeleteRecommendations(ctx context.Context, userID uint64, recType *domain.RecommendationType) error {
	db := r.db.WithContext(ctx).Where("user_id = ?", userID)
	if recType != nil {
		db = db.Where("recommendation_type = ?", *recType)
	}
	return db.Delete(&RecommendationModel{}).Error
}

// --- 用户偏好 (UserPreference methods) ---

func (r *recommendationRepository) SaveUserPreference(ctx context.Context, pref *domain.UserPreference) error {
	model := toUserPreferenceModel(pref)
	if model == nil {
		return nil
	}
	if err := r.db.WithContext(ctx).Save(model).Error; err != nil {
		return err
	}
	*pref = *toUserPreference(model)
	return nil
}

func (r *recommendationRepository) GetUserPreference(ctx context.Context, userID uint64) (*domain.UserPreference, error) {
	var model UserPreferenceModel
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toUserPreference(&model), nil
}

// --- 商品相似度 (ProductSimilarity methods) ---

func (r *recommendationRepository) SaveProductSimilarity(ctx context.Context, sim *domain.ProductSimilarity) error {
	model := toProductSimilarityModel(sim)
	if model == nil {
		return nil
	}
	if err := r.db.WithContext(ctx).Save(model).Error; err != nil {
		return err
	}
	*sim = *toProductSimilarity(model)
	return nil
}

func (r *recommendationRepository) ListSimilarProducts(ctx context.Context, productID uint64, limit int) ([]*domain.ProductSimilarity, error) {
	var list []*ProductSimilarityModel
	if limit <= 0 {
		limit = 10
	}
	if err := r.db.WithContext(ctx).
		Where("product_id = ?", productID).
		Order("similarity desc").
		Limit(limit).
		Find(&list).Error; err != nil {
		return nil, err
	}
	results := make([]*domain.ProductSimilarity, 0, len(list))
	for _, model := range list {
		results = append(results, toProductSimilarity(model))
	}
	return results, nil
}

// --- 用户行为 (UserBehavior methods) ---

func (r *recommendationRepository) SaveUserBehavior(ctx context.Context, behavior *domain.UserBehavior) error {
	model := toUserBehaviorModel(behavior)
	if model == nil {
		return nil
	}
	if err := r.db.WithContext(ctx).Save(model).Error; err != nil {
		return err
	}
	*behavior = *toUserBehavior(model)
	return nil
}

func (r *recommendationRepository) ListUserBehaviors(ctx context.Context, userID uint64, limit int) ([]*domain.UserBehavior, error) {
	var list []*UserBehaviorModel
	if limit <= 0 {
		limit = 10
	}
	if err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("timestamp desc").
		Limit(limit).
		Find(&list).Error; err != nil {
		return nil, err
	}
	results := make([]*domain.UserBehavior, 0, len(list))
	for _, model := range list {
		results = append(results, toUserBehavior(model))
	}
	return results, nil
}

func (r *recommendationRepository) GetRecentBehaviors(ctx context.Context, limit int) ([]*domain.UserBehavior, error) {
	var list []*UserBehaviorModel
	if limit <= 0 {
		limit = 100
	}
	if err := r.db.WithContext(ctx).
		Order("timestamp desc").
		Limit(limit).
		Find(&list).Error; err != nil {
		return nil, err
	}
	results := make([]*domain.UserBehavior, 0, len(list))
	for _, model := range list {
		results = append(results, toUserBehavior(model))
	}
	return results, nil
}
