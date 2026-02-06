// 生成摘要：定义评论搜索仓储接口（Elasticsearch），用于分页与过滤查询。
// 假设：ES 索引字段与 domain.Review 的 JSON 映射一致，created_at 可用于排序。
package domain

import "context"

// ReviewSearchRepository 定义评论搜索仓储接口。
type ReviewSearchRepository interface {
	// Index 将评论写入搜索索引。
	Index(ctx context.Context, review *Review) error
	// Delete 从索引中删除评论。
	Delete(ctx context.Context, reviewID uint64) error
	// Search 按条件检索评论（支持商品/用户/状态过滤、分页）。
	Search(ctx context.Context, productID *uint64, userID *uint64, status *ReviewStatus, offset, limit int, sortBy string) ([]*Review, int64, error)
}
