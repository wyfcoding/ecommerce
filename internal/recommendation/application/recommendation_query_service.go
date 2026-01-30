package application

import (
	"context"

	"github.com/wyfcoding/ecommerce/internal/recommendation/domain"
)

// RecommendationQuery 处理推荐模块的查询操作。
type RecommendationQuery struct {
	repo domain.RecommendationRepository
}

// NewRecommendationQuery 创建并返回一个新的 RecommendationQuery 实例。
func NewRecommendationQuery(repo domain.RecommendationRepository) *RecommendationQuery {
	return &RecommendationQuery{repo: repo}
}

// GetUserRecommendations 获取指定用户的推荐列表。
func (q *RecommendationQuery) GetUserRecommendations(ctx context.Context, userID uint64, recType *domain.RecommendationType, limit int) ([]*domain.Recommendation, error) {
	return q.repo.ListRecommendations(ctx, userID, recType, limit)
}

// GetUserPreference 获取用户的个性化偏好。
func (q *RecommendationQuery) GetUserPreference(ctx context.Context, userID uint64) (*domain.UserPreference, error) {
	return q.repo.GetUserPreference(ctx, userID)
}

// GetSimilarProducts 获取相似商品推荐。
func (q *RecommendationQuery) GetSimilarProducts(ctx context.Context, productID uint64, limit int) ([]*domain.ProductSimilarity, error) {
	return q.repo.ListSimilarProducts(ctx, productID, limit)
}

// GetUserBehaviors 获取用户的行为记录。
func (q *RecommendationQuery) GetUserBehaviors(ctx context.Context, userID uint64, limit int) ([]*domain.UserBehavior, error) {
	return q.repo.ListUserBehaviors(ctx, userID, limit)
}
