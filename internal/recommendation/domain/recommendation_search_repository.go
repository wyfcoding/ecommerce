// 生成摘要：定义推荐搜索仓储接口（Elasticsearch），用于分页与过滤查询。
// 假设：ES 索引可按 user_id、recommendation_type 过滤并按 score 排序。
package domain

import (
	"context"
)

// RecommendationSearchRepository 定义推荐搜索仓储接口。
type RecommendationSearchRepository interface {
	// Index 将推荐写入搜索索引。
	Index(ctx context.Context, rec *Recommendation) error
	// Delete 从索引中删除推荐。
	Delete(ctx context.Context, documentID string) error
	// DeleteByUserAndType 删除指定用户与类型的推荐索引。
	DeleteByUserAndType(ctx context.Context, userID uint64, recType *RecommendationType) error
	// Search 按条件检索推荐（支持用户与类型过滤、分页）。
	Search(ctx context.Context, userID uint64, recType *RecommendationType, offset, limit int) ([]*Recommendation, int64, error)
}
