package application

import (
	"context"

	"github.com/wyfcoding/ecommerce/internal/wishlist/domain"
)

// WishlistService 结构体定义了收藏夹管理相关的应用服务（外观模式）。
// 它将业务逻辑委托给 WishlistManager 和 WishlistQuery 处理。
type WishlistService struct {
	manager *WishlistManager
	query   *WishlistQuery
}

// NewWishlistService 创建并返回一个新的 WishlistService 实例。
func NewWishlistService(manager *WishlistManager, query *WishlistQuery) *WishlistService {
	return &WishlistService{
		manager: manager,
		query:   query,
	}
}

// Add 将商品添加到用户的收藏夹。
func (s *WishlistService) Add(ctx context.Context, userID, productID, skuID uint64, productName, skuName string, price uint64, imageURL string) (*domain.Wishlist, error) {
	return s.manager.AddToWishlist(ctx, userID, productID, skuID, productName, skuName, imageURL, price)
}

// Remove 将商品从用户的收藏夹中移除。
func (s *WishlistService) Remove(ctx context.Context, userID, id uint64) error {
	// Note: In infrastructure, we have Delete(userID, id) and DeleteByProduct(userID, skuID).
	// Here id is the wishlist entry primary key.
	return s.manager.repo.Delete(ctx, userID, id)
}

// RemoveByProduct 将指定商品从用户的收藏夹中移除。
func (s *WishlistService) RemoveByProduct(ctx context.Context, userID, skuID uint64) error {
	return s.manager.RemoveFromWishlist(ctx, userID, skuID)
}

// List 获取指定用户的收藏夹列表。
func (s *WishlistService) List(ctx context.Context, userID uint64, page, pageSize int) ([]*domain.Wishlist, int64, error) {
	return s.query.GetWishlist(ctx, userID, page, pageSize)
}

// CheckStatus 检查指定商品（SKU）是否已在用户的收藏夹中。
func (s *WishlistService) CheckStatus(ctx context.Context, userID, skuID uint64) (bool, error) {
	return s.query.IsInWishlist(ctx, userID, skuID)
}

// Clear 清空指定用户的收藏夹。
func (s *WishlistService) Clear(ctx context.Context, userID uint64) error {
	return s.manager.ClearWishlist(ctx, userID)
}
