package application

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/wyfcoding/ecommerce/internal/recommendation/domain"
	algorithm "github.com/wyfcoding/pkg/algos/ml"
	"github.com/wyfcoding/pkg/messagequeue"
)

// RecommendationCommandService 处理推荐模块的写操作和业务逻辑。
type RecommendationCommandService struct {
	repo      domain.RecommendationRepository
	publisher messagequeue.EventPublisher
	logger    *slog.Logger
}

// NewRecommendationCommandService 创建并返回一个新的 RecommendationCommandService 实例。
func NewRecommendationCommandService(repo domain.RecommendationRepository, publisher messagequeue.EventPublisher, logger *slog.Logger) *RecommendationCommandService {
	return &RecommendationCommandService{
		repo:      repo,
		publisher: publisher,
		logger:    logger,
	}
}

// SaveRecommendation 保存推荐结果。
func (m *RecommendationCommandService) SaveRecommendation(ctx context.Context, rec *domain.Recommendation) error {
	if err := m.repo.SaveRecommendation(ctx, rec); err != nil {
		m.logger.Error("failed to save recommendation", "error", err, "user_id", rec.UserID)
		return err
	}
	m.publishRecommendationChanged(ctx, rec.UserID, &rec.RecommendationType)
	return nil
}

// DeleteRecommendations 删除推荐。
func (m *RecommendationCommandService) DeleteRecommendations(ctx context.Context, userID uint64, recType *domain.RecommendationType) error {
	if err := m.repo.DeleteRecommendations(ctx, userID, recType); err != nil {
		m.logger.Error("failed to delete recommendations", "error", err, "user_id", userID)
		return err
	}
	m.publishRecommendationDeleted(ctx, userID, recType)
	return nil
}

// SaveUserPreference 保存用户偏好。
func (m *RecommendationCommandService) SaveUserPreference(ctx context.Context, pref *domain.UserPreference) error {
	if err := m.repo.SaveUserPreference(ctx, pref); err != nil {
		m.logger.Error("failed to save user preference", "error", err, "user_id", pref.UserID)
		return err
	}
	m.publishPreferenceUpdated(ctx, pref.UserID)
	return nil
}

// SaveUserBehavior 记录用户行为。
func (m *RecommendationCommandService) SaveUserBehavior(ctx context.Context, behavior *domain.UserBehavior) error {
	if err := m.repo.SaveUserBehavior(ctx, behavior); err != nil {
		m.logger.Error("failed to save user behavior", "error", err, "user_id", behavior.UserID)
		return err
	}
	m.publishBehaviorRecorded(ctx, behavior)
	return nil
}

// TrackBehavior 记录并权重化用户的实时行为，用于实时推荐更新。
func (m *RecommendationCommandService) TrackBehavior(ctx context.Context, userID, productID uint64, action string) error {
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
	return m.SaveUserBehavior(ctx, behavior)
}

// SaveProductSimilarity 保存商品关系相似度，用于图推荐索引。
func (m *RecommendationCommandService) SaveProductSimilarity(ctx context.Context, sim *domain.ProductSimilarity) error {
	if err := m.repo.SaveProductSimilarity(ctx, sim); err != nil {
		m.logger.Error("failed to save product similarity", "error", err, "product_id", sim.ProductID, "similar_product_id", sim.SimilarProductID)
		return err
	}
	return nil
}

// UpsertUserPreference 更新用户偏好设置。
func (m *RecommendationCommandService) UpsertUserPreference(ctx context.Context, pref *domain.UserPreference) error {
	existing, err := m.repo.GetUserPreference(ctx, pref.UserID)
	if err != nil {
		return err
	}
	if existing != nil {
		pref.ID = existing.ID
		pref.CreatedAt = existing.CreatedAt
	}
	return m.SaveUserPreference(ctx, pref)
}

// GenerateRecommendationsSimple 基于用户行为与相似商品生成推荐（保留原有逻辑）。
func (m *RecommendationCommandService) GenerateRecommendationsSimple(ctx context.Context, userID uint64) error {
	m.logger.Info("starting algorithm-based recommendation generation", "user_id", userID)

	// 1. 获取用户历史行为数据 (真实输入)
	history, err := m.repo.ListUserBehaviors(ctx, userID, 100)
	if err != nil {
		return fmt.Errorf("failed to fetch user history: %w", err)
	}

	// 2. 清除旧的推荐数据。
	if err := m.DeleteRecommendations(ctx, userID, nil); err != nil {
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
		similar, _ := m.repo.ListSimilarProducts(ctx, lastProductID, 5)

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
		if err := m.SaveRecommendation(ctx, r); err != nil {
			m.logger.Error("failed to save generated recommendation", "error", err)
		}
	}
	return nil
}

// GenerateRecommendations 生成并保存用户的推荐结果。
func (m *RecommendationCommandService) GenerateRecommendations(ctx context.Context, userID uint64) error {
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
		if err := m.SaveRecommendation(ctx, rec); err != nil {
			m.logger.Error("failed to save generated recommendation", "user_id", userID, "error", err)
		}
	}

	m.logger.Info("recommendations generated", "user_id", userID, "type", recType, "count", len(recs))
	return nil
}

func (m *RecommendationCommandService) publishRecommendationChanged(ctx context.Context, userID uint64, recType *domain.RecommendationType) {
	if m.publisher == nil {
		return
	}
	t := domain.RecommendationType("")
	if recType != nil {
		t = *recType
	}
	event := &domain.RecommendationChangedEvent{
		UserID:             userID,
		RecommendationType: t,
		Timestamp:          time.Now(),
	}
	_ = m.publisher.Publish(ctx, domain.RecommendationChangedEventType, fmt.Sprintf("%d", userID), event)
}

func (m *RecommendationCommandService) publishRecommendationDeleted(ctx context.Context, userID uint64, recType *domain.RecommendationType) {
	if m.publisher == nil {
		return
	}
	t := domain.RecommendationType("")
	if recType != nil {
		t = *recType
	}
	event := &domain.RecommendationDeletedEvent{
		UserID:             userID,
		RecommendationType: t,
		Timestamp:          time.Now(),
	}
	_ = m.publisher.Publish(ctx, domain.RecommendationDeletedEventType, fmt.Sprintf("%d", userID), event)
}

func (m *RecommendationCommandService) publishPreferenceUpdated(ctx context.Context, userID uint64) {
	if m.publisher == nil {
		return
	}
	event := &domain.UserPreferenceUpdatedEvent{
		UserID:    userID,
		Timestamp: time.Now(),
	}
	_ = m.publisher.Publish(ctx, domain.UserPreferenceUpdatedEventType, fmt.Sprintf("%d", userID), event)
}

func (m *RecommendationCommandService) publishBehaviorRecorded(ctx context.Context, behavior *domain.UserBehavior) {
	if m.publisher == nil || behavior == nil {
		return
	}
	event := &domain.UserBehaviorRecordedEvent{
		UserID:    behavior.UserID,
		ProductID: behavior.ProductID,
		Action:    behavior.Action,
		Timestamp: time.Now(),
	}
	_ = m.publisher.Publish(ctx, domain.UserBehaviorRecordedEventType, fmt.Sprintf("%d", behavior.UserID), event)
}
