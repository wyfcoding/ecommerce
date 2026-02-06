// 生成摘要：定义推荐读模型仓储接口（Redis），用于高频查询。
// 假设：推荐列表以 user_id + type 为主键缓存。
package domain

import "context"

// RecommendationReadRepository 定义推荐读模型的高性能访问接口。
type RecommendationReadRepository interface {
	// SaveRecommendations 保存指定用户的推荐列表。
	SaveRecommendations(ctx context.Context, userID uint64, recType *RecommendationType, recs []*Recommendation) error
	// GetRecommendations 获取指定用户的推荐列表。
	GetRecommendations(ctx context.Context, userID uint64, recType *RecommendationType, limit int) ([]*Recommendation, error)
	// DeleteRecommendations 删除指定用户的推荐列表缓存。
	DeleteRecommendations(ctx context.Context, userID uint64, recType *RecommendationType) error
}
