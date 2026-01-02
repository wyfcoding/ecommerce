package application

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/wyfcoding/ecommerce/internal/recommendation/domain"
)

// RecommendationService 结构体定义了推荐系统相关的应用服务 (外观模式)。
// 它协调 RecommendationManager 和 RecommendationQuery 处理用户推荐获取、追踪、偏好更新和推荐生成。
type RecommendationService struct {
	manager *RecommendationManager
	query   *RecommendationQuery
	logger  *slog.Logger
}

// NewRecommendationService 创建并返回一个新的 RecommendationService 实例。
func NewRecommendationService(manager *RecommendationManager, query *RecommendationQuery, logger *slog.Logger) *RecommendationService {
	return &RecommendationService{
		manager: manager,
		query:   query,
		logger:  logger,
	}
}

// GetRecommendations 获取指定用户ID的个性化推荐商品列表。
func (s *RecommendationService) GetRecommendations(ctx context.Context, userID uint64, recType string, limit int) ([]*domain.Recommendation, error) {
	var t *domain.RecommendationType
	if recType != "" {
		rt := domain.RecommendationType(recType)
		t = &rt
	}
	return s.query.GetUserRecommendations(ctx, userID, t, limit)
}

// TrackBehavior 记录并权重化用户的实时行为，用于实时推荐更新。
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

// UpdateUserPreference 更新用户的主动偏好设置或标签信息。
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

// GetSimilarProducts 基于商品属性或协同过滤获取相似商品。
func (s *RecommendationService) GetSimilarProducts(ctx context.Context, productID uint64, limit int) ([]*domain.ProductSimilarity, error) {
	return s.query.GetSimilarProducts(ctx, productID, limit)
}

// GenerateRecommendations 核心算法流程：为用户生成并缓存新的推荐商品列表。
func (s *RecommendationService) GenerateRecommendations(ctx context.Context, userID uint64) error {
	s.logger.Info("starting algorithm-based recommendation generation", "user_id", userID)

	// 1. 获取用户历史行为数据 (真实输入)
	history, err := s.query.GetUserBehaviors(ctx, userID, 100)
	if err != nil {
		return fmt.Errorf("failed to fetch user history: %w", err)
	}

	// 2. 清除旧的推荐数据。
	if err := s.manager.DeleteRecommendations(ctx, userID, nil); err != nil {
		return err
	}

	// 3. 真实化生成：基于用户行为的热门与个性化混合推荐
	var recs []*domain.Recommendation

	if len(history) == 0 {
		// 无历史行为，回退到热门推荐 (假设 ID 5001)
		recs = append(recs, &domain.Recommendation{
			UserID:             userID,
			RecommendationType: domain.RecommendationTypeHot,
			ProductID:          5001,
			Score:              0.8,
			Reason:             "Trending globally",
		})
	} else {
		// 基于最近一次行为进行个性化关联
		lastProductID := history[0].ProductID
		similar, _ := s.query.GetSimilarProducts(ctx, lastProductID, 5)

		for _, sim := range similar {
			recs = append(recs, &domain.Recommendation{
				UserID:             userID,
				RecommendationType: domain.RecommendationTypePersonalized,
				ProductID:          sim.SimilarProductID,
				Score:              sim.Similarity,
				Reason:             fmt.Sprintf("Similar to item you %s", history[0].Action),
			})
		}
	}

	// 4. 保存新生成的推荐数据。
	for _, r := range recs {
		if err := s.manager.SaveRecommendation(ctx, r); err != nil {
			s.logger.Error("failed to save generated recommendation", "error", err)
		}
	}
	return nil
}
