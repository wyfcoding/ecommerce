package application

import (
	"context"

	"github.com/wyfcoding/ecommerce/internal/wishlist/domain"
)

// WishlistQueryService 处理收藏夹模块的查询操作。
type WishlistQueryService struct {
	repo       domain.WishlistRepository
	readRepo   domain.WishlistReadRepository
	searchRepo domain.WishlistSearchRepository
}

// NewWishlistQueryService 创建并返回一个新的 WishlistQueryService 实例。
func NewWishlistQueryService(repo domain.WishlistRepository, readRepo domain.WishlistReadRepository, searchRepo domain.WishlistSearchRepository) *WishlistQueryService {
	return &WishlistQueryService{repo: repo, readRepo: readRepo, searchRepo: searchRepo}
}

// GetWishlist 获取指定用户的收藏夹列表。
func (q *WishlistQueryService) GetWishlist(ctx context.Context, userID uint64, page, pageSize int) ([]*domain.Wishlist, int64, error) {
	offset := (page - 1) * pageSize
	if q.readRepo != nil {
		if list, total, err := q.readRepo.List(ctx, userID, offset, pageSize); err == nil && list != nil {
			return list, total, nil
		}
	}
	if q.searchRepo != nil {
		if list, total, err := q.searchRepo.Search(ctx, userID, offset, pageSize); err == nil && list != nil {
			return list, total, nil
		}
	}
	return q.repo.List(ctx, userID, offset, pageSize)
}

// IsInWishlist 检查特定商品（SKU）是否在用户的收藏夹中。
func (q *WishlistQueryService) IsInWishlist(ctx context.Context, userID, skuID uint64) (bool, error) {
	if q.readRepo != nil {
		if item, err := q.readRepo.Get(ctx, userID, skuID); err == nil && item != nil {
			return true, nil
		}
	}
	item, err := q.repo.Get(ctx, userID, skuID)
	if err != nil {
		return false, err
	}
	return item != nil, nil
}
