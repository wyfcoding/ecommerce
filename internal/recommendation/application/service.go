package application

import (
	"context"
	"ecommerce/internal/recommendation/domain/entity"
	"ecommerce/internal/recommendation/domain/repository"
	"time"

	"log/slog"
)

type RecommendationService struct {
	repo   repository.RecommendationRepository
	logger *slog.Logger
}

func NewRecommendationService(repo repository.RecommendationRepository, logger *slog.Logger) *RecommendationService {
	return &RecommendationService{
		repo:   repo,
		logger: logger,
	}
}

// GetRecommendations 获取推荐
func (s *RecommendationService) GetRecommendations(ctx context.Context, userID uint64, recType string, limit int) ([]*entity.Recommendation, error) {
	var t *entity.RecommendationType
	if recType != "" {
		rt := entity.RecommendationType(recType)
		t = &rt
	}
	return s.repo.ListRecommendations(ctx, userID, t, limit)
}

// TrackBehavior 记录用户行为
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

	behavior := &entity.UserBehavior{
		UserID:    userID,
		ProductID: productID,
		Action:    action,
		Weight:    weight,
		Timestamp: time.Now(),
	}

	return s.repo.SaveUserBehavior(ctx, behavior)
}

// UpdateUserPreference 更新用户偏好
func (s *RecommendationService) UpdateUserPreference(ctx context.Context, pref *entity.UserPreference) error {
	// Check if exists
	existing, err := s.repo.GetUserPreference(ctx, pref.UserID)
	if err != nil {
		return err
	}
	if existing != nil {
		pref.ID = existing.ID
		pref.CreatedAt = existing.CreatedAt
	}
	return s.repo.SaveUserPreference(ctx, pref)
}

// GetSimilarProducts 获取相似商品
func (s *RecommendationService) GetSimilarProducts(ctx context.Context, productID uint64, limit int) ([]*entity.ProductSimilarity, error) {
	return s.repo.ListSimilarProducts(ctx, productID, limit)
}

// GenerateRecommendations 生成推荐 (Mock logic for now)
func (s *RecommendationService) GenerateRecommendations(ctx context.Context, userID uint64) error {
	// In a real system, this would invoke a complex algorithm or call an AI model.
	// Here we just mock some data based on recent behavior or random.

	// Clear old recommendations
	if err := s.repo.DeleteRecommendations(ctx, userID, nil); err != nil {
		return err
	}

	// Mock recommendations
	recs := []*entity.Recommendation{
		{
			UserID:             userID,
			RecommendationType: entity.RecommendationTypePersonalized,
			ProductID:          101,
			Score:              0.95,
			Reason:             "Based on your viewing history",
		},
		{
			UserID:             userID,
			RecommendationType: entity.RecommendationTypeHot,
			ProductID:          202,
			Score:              0.88,
			Reason:             "Popular item",
		},
	}

	for _, r := range recs {
		if err := s.repo.SaveRecommendation(ctx, r); err != nil {
			return err
		}
	}

	return nil
}
