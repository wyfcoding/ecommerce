package application

import (
	"context"
	"fmt"
	"time"

	"github.com/wyfcoding/ecommerce/internal/recommendation/domain"
)

// RecommendationService 结构体定义了推荐系统相关的应用服务 (外观模式)。
// 它协调 RecommendationManager 和 RecommendationQuery 处理用户推荐获取、追踪、偏好更新和推荐生成。
type RecommendationService struct {
	manager *RecommendationManager
	query   *RecommendationQuery
}

// NewRecommendationService 创建并返回一个新的 RecommendationService 实例。
func NewRecommendationService(manager *RecommendationManager, query *RecommendationQuery) *RecommendationService {
	return &RecommendationService{
		manager: manager,
		query:   query,
	}
}

// GetRecommendations 获取指定用户ID的推荐列表。
func (s *RecommendationService) GetRecommendations(ctx context.Context, userID uint64, recType string, limit int) ([]*domain.Recommendation, error) {
	var t *domain.RecommendationType
	if recType != "" {
		rt := domain.RecommendationType(recType)
		t = &rt
	}
	return s.query.GetUserRecommendations(ctx, userID, t, limit)
}

// TrackBehavior 记录用户行为。
func (s *RecommendationService) TrackBehavior(ctx context.Context, userID, productID uint64, action string) error {
	weight := 1.0
	switch action {
	case "view":
		weight = 1.0
	case "click":
		weight = 2.0
	case "cart":
		weight = 5.0
	case "buy":
		weight = 10.0
	}

	behavior := &domain.UserBehavior{
		UserID:    userID,
		ProductID: productID,
		Action:    action,
		Weight:    weight,
		Timestamp: time.Now(),
	}

	return s.manager.SaveUserBehavior(ctx, behavior)
}

// UpdateUserPreference 更新用户偏好设置。
func (s *RecommendationService) UpdateUserPreference(ctx context.Context, pref *domain.UserPreference) error {
	existing, err := s.query.GetUserPreference(ctx, pref.UserID)
	if err != nil {
		return err
	}
	if existing != nil {
		pref.ID = existing.ID
		pref.CreatedAt = existing.CreatedAt
	}
	return s.manager.SaveUserPreference(ctx, pref)
}

// GetSimilarProducts 获取相似商品列表。
func (s *RecommendationService) GetSimilarProducts(ctx context.Context, productID uint64, limit int) ([]*domain.ProductSimilarity, error) {
	return s.query.GetSimilarProducts(ctx, productID, limit)
}

// GenerateRecommendations 生成推荐列表 (模拟业务流程)。
func (s *RecommendationService) GenerateRecommendations(ctx context.Context, userID uint64) error {
	// 1. 清除旧的推荐数据。
	if err := s.manager.DeleteRecommendations(ctx, userID, nil); err != nil {
		return err
	}

	// 2. 模拟生成新的推荐数据。
	recs := []*domain.Recommendation{
		{
			UserID:             userID,
			RecommendationType: domain.RecommendationTypePersonalized,
			ProductID:          101,
			Score:              0.95,
			Reason:             "Based on your viewing history",
		},
		{
			UserID:             userID,
			RecommendationType: domain.RecommendationTypeHot,
			ProductID:          202,
			Score:              0.88,
			Reason:             "Popular item",
		},
	}

	// 3. 保存新生成的推荐数据。
	for _, r := range recs {
		if err := s.manager.SaveRecommendation(ctx, r); err != nil {
			return fmt.Errorf("failed to save recommendation: %w", err)
		}
	}
	return nil
}
