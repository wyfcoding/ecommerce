package persistence

import (
	"context"
	"errors" // 导入标准错误处理库。

	"github.com/wyfcoding/ecommerce/internal/recommendation/domain" // 导入推荐领域的领域定义。

	"gorm.io/gorm" // 导入GORM ORM框架。
)

type recommendationRepository struct {
	db *gorm.DB // GORM数据库连接实例。
}

// NewRecommendationRepository 创建并返回一个新的 recommendationRepository 实例。
func NewRecommendationRepository(db *gorm.DB) domain.RecommendationRepository {
	return &recommendationRepository{db: db}
}

// --- 推荐结果 (Recommendation methods) ---

// SaveRecommendation 将推荐结果实体保存到数据库。
func (r *recommendationRepository) SaveRecommendation(ctx context.Context, rec *domain.Recommendation) error {
	return r.db.WithContext(ctx).Save(rec).Error
}

// ListRecommendations 从数据库列出指定用户ID的推荐结果实体，支持通过推荐类型过滤和数量限制。
func (r *recommendationRepository) ListRecommendations(ctx context.Context, userID uint64, recType *domain.RecommendationType, limit int) ([]*domain.Recommendation, error) {
	var list []*domain.Recommendation
	db := r.db.WithContext(ctx).Where("user_id = ?", userID) // 按用户ID过滤。
	if recType != nil {                                      // 如果提供了推荐类型，则按类型过滤。
		db = db.Where("recommendation_type = ?", *recType)
	}
	// 按推荐分数降序排列，并应用数量限制。
	if err := db.Order("score desc").Limit(limit).Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

// DeleteRecommendations 删除指定用户ID和推荐类型（可选）的所有推荐结果。
func (r *recommendationRepository) DeleteRecommendations(ctx context.Context, userID uint64, recType *domain.RecommendationType) error {
	db := r.db.WithContext(ctx).Where("user_id = ?", userID) // 按用户ID过滤。
	if recType != nil {                                      // 如果提供了推荐类型，则按类型过滤。
		db = db.Where("recommendation_type = ?", *recType)
	}
	// 执行删除操作。
	return db.Delete(&domain.Recommendation{}).Error
}

// --- 用户偏好 (UserPreference methods) ---

// SaveUserPreference 将用户偏好实体保存到数据库。
func (r *recommendationRepository) SaveUserPreference(ctx context.Context, pref *domain.UserPreference) error {
	return r.db.WithContext(ctx).Save(pref).Error
}

// GetUserPreference 根据用户ID从数据库获取用户偏好记录。
func (r *recommendationRepository) GetUserPreference(ctx context.Context, userID uint64) (*domain.UserPreference, error) {
	var pref domain.UserPreference
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&pref).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // 如果记录未找到，返回nil。
		}
		return nil, err // 其他错误则返回。
	}
	return &pref, nil
}

// --- 商品相似度 (ProductSimilarity methods) ---

// SaveProductSimilarity 将商品相似度实体保存到数据库。
func (r *recommendationRepository) SaveProductSimilarity(ctx context.Context, sim *domain.ProductSimilarity) error {
	return r.db.WithContext(ctx).Save(sim).Error
}

// ListSimilarProducts 从数据库列出指定商品ID的相似商品实体，支持数量限制。
func (r *recommendationRepository) ListSimilarProducts(ctx context.Context, productID uint64, limit int) ([]*domain.ProductSimilarity, error) {
	var list []*domain.ProductSimilarity
	// 按商品ID过滤，按相似度降序排列，并应用数量限制。
	if err := r.db.WithContext(ctx).Where("product_id = ?", productID).Order("similarity desc").Limit(limit).Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

// --- 用户行为 (UserBehavior methods) ---

// SaveUserBehavior 将用户行为实体保存到数据库。
func (r *recommendationRepository) SaveUserBehavior(ctx context.Context, behavior *domain.UserBehavior) error {
	return r.db.WithContext(ctx).Save(behavior).Error
}

// ListUserBehaviors 从数据库列出指定用户ID的用户行为实体，支持数量限制。
func (r *recommendationRepository) ListUserBehaviors(ctx context.Context, userID uint64, limit int) ([]*domain.UserBehavior, error) {
	var list []*domain.UserBehavior
		return list, nil
	}
	
	// GetRecentBehaviors 获取最近的全站用户行为。
	func (r *recommendationRepository) GetRecentBehaviors(ctx context.Context, limit int) ([]*domain.UserBehavior, error) {
		var list []*domain.UserBehavior
		if err := r.db.WithContext(ctx).Order("timestamp desc").Limit(limit).Find(&list).Error; err != nil {
			return nil, err
		}
		return list, nil
	}
	
