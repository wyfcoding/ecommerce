// 生成摘要：定义收藏夹读模型仓储接口（Redis），用于高频查询。
// 假设：读模型以 user_id + sku_id 为主键缓存，并维护用户维度索引。
package domain

import "context"

// WishlistReadRepository 定义收藏夹读模型的高性能访问接口。
type WishlistReadRepository interface {
	Save(ctx context.Context, item *Wishlist) error
	Delete(ctx context.Context, userID, skuID uint64) error
	Get(ctx context.Context, userID, skuID uint64) (*Wishlist, error)
	List(ctx context.Context, userID uint64, offset, limit int) ([]*Wishlist, int64, error)
	Clear(ctx context.Context, userID uint64) error
}
