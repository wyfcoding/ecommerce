package domain

import (
	"context"
)

// WishlistRepository 是收藏夹模块的仓储接口。
// 它定义了对 Wishlist 实体进行数据持久化操作的契约。
// 仓储接口属于领域层，旨在将领域逻辑与数据存储的实现细节解耦。
type WishlistRepository interface {
	// Save 将收藏夹实体保存到数据存储中。
	// 如果实体已存在，则更新；如果不存在，则创建。
	// ctx: 上下文。
	// wishlist: 待保存的收藏夹实体。
	Save(ctx context.Context, wishlist *Wishlist) error
	// Delete 从数据存储中删除指定用户ID和收藏夹条目ID的记录。
	Delete(ctx context.Context, userID, id uint64) error
	// DeleteByProduct 从数据存储中删除指定用户ID和商品ID（SKUID）的记录。
	DeleteByProduct(ctx context.Context, userID, skuID uint64) error
	// Get 获取指定用户ID和SKUID的收藏夹实体。
	Get(ctx context.Context, userID, skuID uint64) (*Wishlist, error)
	// List 列出指定用户ID的所有收藏夹实体，支持分页。
	List(ctx context.Context, userID uint64, offset, limit int) ([]*Wishlist, int64, error)
	// Count 统计指定用户ID的收藏夹条目数量。
	Count(ctx context.Context, userID uint64) (int64, error)
	// Clear 清空指定用户的收藏夹。
	Clear(ctx context.Context, userID uint64) error
}
