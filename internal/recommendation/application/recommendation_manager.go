package application

import (
	"context"
	"log/slog"

	"github.com/wyfcoding/ecommerce/internal/recommendation/domain"
)

// RecommendationManager 处理推荐模块的写操作和业务逻辑。
type RecommendationManager struct {
	repo   domain.RecommendationRepository
	logger *slog.Logger
}

// NewRecommendationManager 创建并返回一个新的 RecommendationManager 实例。
func NewRecommendationManager(repo domain.RecommendationRepository, logger *slog.Logger) *RecommendationManager {
	return &RecommendationManager{
		repo:   repo,
		logger: logger,
	}
}

// SaveRecommendation 保存推荐结果。
func (m *RecommendationManager) SaveRecommendation(ctx context.Context, rec *domain.Recommendation) error {
	if err := m.repo.SaveRecommendation(ctx, rec); err != nil {
		m.logger.Error("failed to save recommendation", "error", err, "user_id", rec.UserID)
		return err
	}
	return nil
}

// DeleteRecommendations 删除推荐。
func (m *RecommendationManager) DeleteRecommendations(ctx context.Context, userID uint64, recType *domain.RecommendationType) error {
	if err := m.repo.DeleteRecommendations(ctx, userID, recType); err != nil {
		m.logger.Error("failed to delete recommendations", "error", err, "user_id", userID)
		return err
	}
	return nil
}

// SaveUserPreference 保存用户偏好。
func (m *RecommendationManager) SaveUserPreference(ctx context.Context, pref *domain.UserPreference) error {
	if err := m.repo.SaveUserPreference(ctx, pref); err != nil {
		m.logger.Error("failed to save user preference", "error", err, "user_id", pref.UserID)
		return err
	}
	return nil
}

// SaveUserBehavior 记录用户行为。
func (m *RecommendationManager) SaveUserBehavior(ctx context.Context, behavior *domain.UserBehavior) error {
	if err := m.repo.SaveUserBehavior(ctx, behavior); err != nil {
		m.logger.Error("failed to save user behavior", "error", err, "user_id", behavior.UserID)
		return err
	}
	return nil
}
