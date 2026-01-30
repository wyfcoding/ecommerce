package application

import (
	"context"

	"github.com/wyfcoding/ecommerce/internal/recommendation/domain"
)

// RecommendationQueryService 处理推荐模块的查询操作。
type RecommendationQueryService struct {
	repo domain.RecommendationRepository
}

// NewRecommendationQueryService 创建并返回一个新的 RecommendationQueryService 实例。
func NewRecommendationQueryService(repo domain.RecommendationRepository) *RecommendationQueryService {
	return &RecommendationQueryService{repo: repo}
}

// GetUserRecommendations 获取指定用户的推荐列表。
func (q *RecommendationQueryService) GetUserRecommendations(ctx context.Context, userID uint64, recType *domain.RecommendationType, limit int) ([]*domain.Recommendation, error) {
	return q.repo.ListRecommendations(ctx, userID, recType, limit)
}

// GetUserPreference 获取用户的个性化偏好。
func (q *RecommendationQueryService) GetUserPreference(ctx context.Context, userID uint64) (*domain.UserPreference, error) {
	return q.repo.GetUserPreference(ctx, userID)
}

// GetSimilarProducts 获取相似商品推荐。
func (q *RecommendationQueryService) GetSimilarProducts(ctx context.Context, productID uint64, limit int) ([]*domain.ProductSimilarity, error) {
	return q.repo.ListSimilarProducts(ctx, productID, limit)
}

// GetUserBehaviors 获取用户的行为记录。
func (q *RecommendationQueryService) GetUserBehaviors(ctx context.Context, userID uint64, limit int) ([]*domain.UserBehavior, error) {
	return q.repo.ListUserBehaviors(ctx, userID, limit)
}
