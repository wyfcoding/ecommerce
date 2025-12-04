package application

import (
	"context" // 导入标准错误处理库。
	"fmt"     // 导入格式化库。
	"time"    // 导入时间库。

	"github.com/wyfcoding/ecommerce/internal/recommendation/domain/entity"     // 导入推荐领域的实体定义。
	"github.com/wyfcoding/ecommerce/internal/recommendation/domain/repository" // 导入推荐领域的仓储接口。

	"log/slog" // 导入结构化日志库。
)

// RecommendationService 结构体定义了推荐系统相关的应用服务。
// 它协调领域层和基础设施层，处理用户推荐的获取、用户行为追踪、用户偏好更新以及推荐的生成等业务逻辑。
type RecommendationService struct {
	repo   repository.RecommendationRepository // 依赖RecommendationRepository接口，用于数据持久化操作。
	logger *slog.Logger                        // 日志记录器，用于记录服务运行时的信息和错误。
}

// NewRecommendationService 创建并返回一个新的 RecommendationService 实例。
func NewRecommendationService(repo repository.RecommendationRepository, logger *slog.Logger) *RecommendationService {
	return &RecommendationService{
		repo:   repo,
		logger: logger,
	}
}

// GetRecommendations 获取指定用户ID的推荐列表。
// ctx: 上下文。
// userID: 用户的唯一标识符。
// recType: 推荐类型（例如，“personalized”，“hot”），可选。
// limit: 推荐数量限制。
// 返回推荐实体列表和可能发生的错误。
func (s *RecommendationService) GetRecommendations(ctx context.Context, userID uint64, recType string, limit int) ([]*entity.Recommendation, error) {
	var t *entity.RecommendationType // 用于过滤推荐类型。
	if recType != "" {
		rt := entity.RecommendationType(recType)
		t = &rt
	}
	return s.repo.ListRecommendations(ctx, userID, t, limit)
}

// TrackBehavior 记录用户行为。
// ctx: 上下文。
// userID: 用户ID。
// productID: 商品ID。
// action: 行为类型（例如，“view”，“click”，“cart”，“buy”）。
// 返回可能发生的错误。
func (s *RecommendationService) TrackBehavior(ctx context.Context, userID, productID uint64, action string) error {
	// 根据行为类型赋予不同的权重，以反映行为的重要性。
	weight := 1.0
	switch action {
	case "view":
		weight = 1.0 // 浏览行为权重。
	case "click":
		weight = 2.0 // 点击行为权重。
	case "cart":
		weight = 5.0 // 加入购物车行为权重。
	case "buy":
		weight = 10.0 // 购买行为权重。
	}

	behavior := &entity.UserBehavior{
		UserID:    userID,
		ProductID: productID,
		Action:    action,
		Weight:    weight,
		Timestamp: time.Now(), // 记录行为发生时间。
	}

	return s.repo.SaveUserBehavior(ctx, behavior)
}

// UpdateUserPreference 更新用户偏好设置。
// ctx: 上下文。
// pref: 待更新的UserPreference实体。
// 返回可能发生的错误。
func (s *RecommendationService) UpdateUserPreference(ctx context.Context, pref *entity.UserPreference) error {
	// 检查用户偏好设置是否已存在。
	existing, err := s.repo.GetUserPreference(ctx, pref.UserID)
	if err != nil {
		return err
	}
	// 如果存在，则更新现有记录的ID和创建时间，以确保正确更新。
	if existing != nil {
		pref.ID = existing.ID
		pref.CreatedAt = existing.CreatedAt
	}
	return s.repo.SaveUserPreference(ctx, pref)
}

// GetSimilarProducts 获取相似商品列表。
// ctx: 上下文。
// productID: 商品ID。
// limit: 限制相似商品数量。
// 返回相似商品列表和可能发生的错误。
func (s *RecommendationService) GetSimilarProducts(ctx context.Context, productID uint64, limit int) ([]*entity.ProductSimilarity, error) {
	return s.repo.ListSimilarProducts(ctx, productID, limit)
}

// GenerateRecommendations 生成推荐列表。
// ctx: 上下文。
// userID: 用户ID。
// 返回可能发生的错误。
func (s *RecommendationService) GenerateRecommendations(ctx context.Context, userID uint64) error {
	// TODO: 在实际系统中，这里会调用复杂的推荐算法或AI模型，根据用户的历史行为、偏好、商品特征等生成个性化推荐。
	// 当前实现仅为模拟逻辑，生成一些硬编码的推荐数据。

	// 1. 清除旧的推荐数据。
	if err := s.repo.DeleteRecommendations(ctx, userID, nil); err != nil {
		return err
	}

	// 2. 模拟生成新的推荐数据。
	recs := []*entity.Recommendation{
		{
			UserID:             userID,
			RecommendationType: entity.RecommendationTypePersonalized,
			ProductID:          101, // 模拟推荐商品ID。
			Score:              0.95,
			Reason:             "Based on your viewing history", // 推荐理由。
		},
		{
			UserID:             userID,
			RecommendationType: entity.RecommendationTypeHot,
			ProductID:          202, // 模拟推荐商品ID。
			Score:              0.88,
			Reason:             "Popular item",
		},
	}

	// 3. 保存新生成的推荐数据。
	for _, r := range recs {
		if err := s.repo.SaveRecommendation(ctx, r); err != nil {
			s.logger.ErrorContext(ctx, fmt.Sprintf("failed to save recommendation for user %d", userID), "product_id", r.ProductID, "error", err)
			return err
		}
	}
	s.logger.InfoContext(ctx, fmt.Sprintf("generated %d recommendations for user %d", len(recs), userID))
	return nil
}
