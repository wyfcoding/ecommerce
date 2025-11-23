package repository

import (
	"context"
	"ecommerce/internal/recommendation/domain/entity"
)

// RecommendationRepository 推荐仓储接口
type RecommendationRepository interface {
	// 推荐结果
	SaveRecommendation(ctx context.Context, rec *entity.Recommendation) error
	ListRecommendations(ctx context.Context, userID uint64, recType *entity.RecommendationType, limit int) ([]*entity.Recommendation, error)
	DeleteRecommendations(ctx context.Context, userID uint64, recType *entity.RecommendationType) error

	// 用户偏好
	SaveUserPreference(ctx context.Context, pref *entity.UserPreference) error
	GetUserPreference(ctx context.Context, userID uint64) (*entity.UserPreference, error)

	// 商品相似度
	SaveProductSimilarity(ctx context.Context, sim *entity.ProductSimilarity) error
	ListSimilarProducts(ctx context.Context, productID uint64, limit int) ([]*entity.ProductSimilarity, error)

	// 用户行为
	SaveUserBehavior(ctx context.Context, behavior *entity.UserBehavior) error
	ListUserBehaviors(ctx context.Context, userID uint64, limit int) ([]*entity.UserBehavior, error)
}
