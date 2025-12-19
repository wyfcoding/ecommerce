package application

import (
	"context"

	"github.com/wyfcoding/ecommerce/internal/wishlist/domain"
)

// WishlistQuery 处理收藏夹模块的查询操作。
type WishlistQuery struct {
	repo domain.WishlistRepository
}

// NewWishlistQuery 创建并返回一个新的 WishlistQuery 实例。
func NewWishlistQuery(repo domain.WishlistRepository) *WishlistQuery {
	return &WishlistQuery{repo: repo}
}

// GetWishlist 获取指定用户的收藏夹列表。
func (q *WishlistQuery) GetWishlist(ctx context.Context, userID uint64, page, pageSize int) ([]*domain.Wishlist, int64, error) {
	offset := (page - 1) * pageSize
	return q.repo.List(ctx, userID, offset, pageSize)
}

// IsInWishlist 检查特定商品（SKU）是否在用户的收藏夹中。
func (q *WishlistQuery) IsInWishlist(ctx context.Context, userID, skuID uint64) (bool, error) {
	item, err := q.repo.Get(ctx, userID, skuID)
	if err != nil {
		return false, err
	}
	return item != nil, nil
}
