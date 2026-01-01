package domain

import (
	"context"
)

// RecommendationRepository 是推荐模块的仓储接口。
// 它定义了对推荐结果、用户偏好、商品相似度和用户行为实体进行数据持久化操作的契约。
// 仓储接口属于领域层，旨在将领域逻辑与数据存储的实现细节解耦。
type RecommendationRepository interface {
	// --- 推荐结果 (Recommendation methods) ---

	// SaveRecommendation 将推荐结果实体保存到数据存储中。
	// ctx: 上下文。
	// rec: 待保存的推荐结果实体。
	SaveRecommendation(ctx context.Context, rec *Recommendation) error
	// ListRecommendations 列出指定用户ID的推荐结果实体，支持通过推荐类型过滤和数量限制。
	ListRecommendations(ctx context.Context, userID uint64, recType *RecommendationType, limit int) ([]*Recommendation, error)
	// DeleteRecommendations 删除指定用户ID和推荐类型（可选）的所有推荐结果。
	DeleteRecommendations(ctx context.Context, userID uint64, recType *RecommendationType) error

	// --- 用户偏好 (UserPreference methods) ---

	// SaveUserPreference 将用户偏好实体保存到数据存储中。
	SaveUserPreference(ctx context.Context, pref *UserPreference) error
	// GetUserPreference 根据用户ID获取用户偏好实体。
	GetUserPreference(ctx context.Context, userID uint64) (*UserPreference, error)

	// --- 商品相似度 (ProductSimilarity methods) ---

	// SaveProductSimilarity 将商品相似度实体保存到数据存储中。
	SaveProductSimilarity(ctx context.Context, sim *ProductSimilarity) error
	// ListSimilarProducts 列出指定商品ID的相似商品实体，支持数量限制。
	ListSimilarProducts(ctx context.Context, productID uint64, limit int) ([]*ProductSimilarity, error)

	// --- 用户行为 (UserBehavior methods) ---

	// SaveUserBehavior 将用户行为实体保存到数据存储中。
	SaveUserBehavior(ctx context.Context, behavior *UserBehavior) error
	// ListUserBehaviors 列出指定用户ID的用户行为实体，支持数量限制。
	ListUserBehaviors(ctx context.Context, userID uint64, limit int) ([]*UserBehavior, error)
	// GetRecentBehaviors 获取最近的全站用户行为（用于构建协同过滤矩阵）。
	GetRecentBehaviors(ctx context.Context, limit int) ([]*UserBehavior, error)
}
