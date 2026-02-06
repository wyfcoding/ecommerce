// 生成摘要：定义收藏夹搜索仓储接口（Elasticsearch），用于分页与过滤查询。
// 假设：ES 索引可按 user_id 过滤并按 created_at 排序。
package domain

import "context"

// WishlistSearchRepository 定义收藏夹搜索仓储接口。
type WishlistSearchRepository interface {
	Index(ctx context.Context, item *Wishlist) error
	Delete(ctx context.Context, documentID string) error
	DeleteByUser(ctx context.Context, userID uint64) error
	Search(ctx context.Context, userID uint64, offset, limit int) ([]*Wishlist, int64, error)
}
