package application

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/wyfcoding/ecommerce/internal/wishlist/domain"
)

// WishlistManager 处理收藏夹模块的写操作和核心业务逻辑。
type WishlistManager struct {
	repo   domain.WishlistRepository
	logger *slog.Logger
}

// NewWishlistManager 创建并返回一个新的 WishlistManager 实例。
func NewWishlistManager(repo domain.WishlistRepository, logger *slog.Logger) *WishlistManager {
	return &WishlistManager{
		repo:   repo,
		logger: logger,
	}
}

// AddToWishlist 将商品添加到收藏夹。
func (m *WishlistManager) AddToWishlist(ctx context.Context, userID, productID, skuID uint64, productName, skuName, imageURL string, price uint64) (*domain.Wishlist, error) {
	// 检查是否已存在。
	existing, err := m.repo.Get(ctx, userID, skuID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return existing, nil // Return existing if already there.
	}

	// 检查收藏夹条目数量限制（假设限制为100）。
	count, err := m.repo.Count(ctx, userID)
	if err != nil {
		return nil, err
	}
	if count >= 100 {
		return nil, fmt.Errorf("wishlist is full (max 100 items)")
	}

	item := &domain.Wishlist{
		UserID:      userID,
		ProductID:   productID,
		SkuID:       skuID,
		ProductName: productName,
		SkuName:     skuName,
		Price:       price,
		ImageURL:    imageURL,
	}

	if err := m.repo.Save(ctx, item); err != nil {
		m.logger.Error("failed to add to wishlist", "error", err, "user_id", userID, "sku_id", skuID)
		return nil, err
	}

	return item, nil
}

// RemoveFromWishlist 从收藏夹中移除指定商品。
func (m *WishlistManager) RemoveFromWishlist(ctx context.Context, userID, skuID uint64) error {
	if err := m.repo.DeleteByProduct(ctx, userID, skuID); err != nil {
		m.logger.Error("failed to remove from wishlist", "error", err, "user_id", userID, "sku_id", skuID)
		return err
	}
	return nil
}

// ClearWishlist 清空用户的收藏夹。
func (m *WishlistManager) ClearWishlist(ctx context.Context, userID uint64) error {
	if err := m.repo.Clear(ctx, userID); err != nil {
		m.logger.Error("failed to clear wishlist", "error", err, "user_id", userID)
		return err
	}
	return nil
}
