// 生成摘要：定义购物车搜索仓储接口（Elasticsearch）。
// 假设：购物车以 user_id 作为检索主键。
package domain

import "context"

// CartSearchRepository 定义购物车搜索的访问接口。
type CartSearchRepository interface {
	// Index 保存或更新购物车搜索文档。
	Index(ctx context.Context, cart *Cart) error
	// Delete 删除购物车搜索文档。
	Delete(ctx context.Context, userID uint64) error
	// Search 按用户检索购物车（保留扩展空间）。
	Search(ctx context.Context, userID uint64) (*Cart, error)
}
