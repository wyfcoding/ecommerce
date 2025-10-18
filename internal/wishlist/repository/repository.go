package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"ecommerce/internal/wishlist/model"
)

// WishlistRepository 定义了心愿单数据仓库的接口
type WishlistRepository interface {
	AddItem(ctx context.Context, userID, productID uint) (*model.WishlistItem, error)
	RemoveItem(ctx context.Context, userID, productID uint) error
	ListItemsByUserID(ctx context.Context, userID uint) ([]model.WishlistItem, error)
	ItemExists(ctx context.Context, userID, productID uint) (bool, error)
}

// wishlistRepository 是接口的具体实现
type wishlistRepository struct {
	db *gorm.DB
}

// NewWishlistRepository 创建一个新的 wishlistRepository 实例
func NewWishlistRepository(db *gorm.DB) WishlistRepository {
	return &wishlistRepository{db: db}
}

// AddItem 向用户的心愿单中添加一个商品
func (r *wishlistRepository) AddItem(ctx context.Context, userID, productID uint) (*model.WishlistItem, error) {
	item := &model.WishlistItem{
		UserID:    userID,
		ProductID: productID,
	}
	// 由于有唯一索引，如果记录已存在，Create 会返回错误
	if err := r.db.WithContext(ctx).Create(item).Error; err != nil {
		return nil, fmt.Errorf("数据库添加心愿单项目失败: %w", err)
	}
	return item, nil
}

// RemoveItem 从用户的心愿单中移除一个商品
func (r *wishlistRepository) RemoveItem(ctx context.Context, userID, productID uint) error {
	result := r.db.WithContext(ctx).Where("user_id = ? AND product_id = ?", userID, productID).Delete(&model.WishlistItem{})
	if result.Error != nil {
		return fmt.Errorf("数据库移除心愿单项目失败: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("该商品不在心愿单中")
	}
	return nil
}

// ListItemsByUserID 列出某个用户心愿单中的所有项目
func (r *wishlistRepository) ListItemsByUserID(ctx context.Context, userID uint) ([]model.WishlistItem, error) {
	var items []model.WishlistItem
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Order("added_at desc").Find(&items).Error; err != nil {
		return nil, fmt.Errorf("数据库列出心愿单失败: %w", err)
	}
	return items, nil
}

// ItemExists 检查某个商品是否已在用户的心愿单中
func (r *wishlistRepository) ItemExists(ctx context.Context, userID, productID uint) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.WishlistItem{}).Where("user_id = ? AND product_id = ?", userID, productID).Count(&count).Error
	if err != nil {
		return false, fmt.Errorf("数据库检查心愿单项目存在性失败: %w", err)
	}
	return count > 0, nil
}
