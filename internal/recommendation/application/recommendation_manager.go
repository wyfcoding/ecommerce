package application

import (
	"context"
	"log/slog"

	"github.com/wyfcoding/ecommerce/internal/recommendation/domain"
	"github.com/wyfcoding/pkg/algorithm"
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

// GenerateRecommendations 生成并保存用户的推荐结果。
func (m *RecommendationManager) GenerateRecommendations(ctx context.Context, userID uint64) error {
	// 1. 获取数据
	userBehaviors, err := m.repo.ListUserBehaviors(ctx, userID, 100)
	if err != nil {
		return err
	}

	globalBehaviors, err := m.repo.GetRecentBehaviors(ctx, 1000)
	if err != nil {
		return err
	}

	// 2. 初始化推荐引擎并加载数据
	engine := algorithm.NewRecommendationEngine()

	mapScore := func(action string) float64 {
		switch action {
		case "buy":
			return 5.0
		case "cart":
			return 3.0
		case "click":
			return 2.0
		case "view":
			return 1.0
		default:
			return 1.0
		}
	}

	// 加载全局数据构建矩阵
	for _, b := range globalBehaviors {
		score := mapScore(b.Action)
		engine.AddRating(b.UserID, b.ProductID, score)
		switch b.Action {
		case "view":
			engine.AddView(b.ProductID)
		case "buy":
			engine.AddSale(b.ProductID)
		}
	}

	// 确保当前用户数据也在其中 (GetRecentBehaviors 可能已包含，但不一定全)
	for _, b := range userBehaviors {
		engine.AddRating(b.UserID, b.ProductID, mapScore(b.Action))
	}

	// 3. 生成推荐 (真实算法分值提取)
	type recItem struct {
		pid   uint64
		score float64
	}
	var recs []recItem
	var recType domain.RecommendationType
	var reason string

	if len(userBehaviors) > 0 {
		// 真实化执行：获取带权重的推荐列表 (假设算法包已支持返回带分值的对象)
		// 这里暂存结果
		items := engine.RecommendWithScores(userID, 10)
		for pid, score := range items {
			recs = append(recs, recItem{pid, score})
		}
		recType = domain.RecommendationTypePersonalized
		reason = "Personalized for you"
	}

	if len(recs) == 0 {
		// 冷启动：热门推荐 (按销售额/点击率权重)
		hot := engine.HotItems(10, 24)
		for i, pid := range hot {
			recs = append(recs, recItem{pid, 1.0 - float64(i)*0.05})
		}
		recType = domain.RecommendationTypeHot
		reason = "Trending now"
	}

	// 4. 保存结果
	if err := m.DeleteRecommendations(ctx, userID, &recType); err != nil {
		return err
	}

	for _, item := range recs {
		rec := &domain.Recommendation{
			UserID:             userID,
			RecommendationType: recType,
			ProductID:          item.pid,
			Score:              item.score,
			Reason:             reason,
		}
		if err := m.repo.SaveRecommendation(ctx, rec); err != nil {
			m.logger.Error("failed to save generated recommendation", "user_id", userID, "error", err)
		}
	}

	m.logger.Info("recommendations generated", "user_id", userID, "type", recType, "count", len(recs))
	return nil
}
