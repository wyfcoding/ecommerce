package mysql

import (
	"context"
	"errors"

	"github.com/wyfcoding/ecommerce/internal/wishlist/domain"
	"gorm.io/gorm"
)

type wishlistRepository struct {
	db *gorm.DB
}

// NewWishlistRepository 创建并返回一个新的 WishlistRepository 实例。
func NewWishlistRepository(db *gorm.DB) domain.WishlistRepository {
	return &wishlistRepository{db: db}
}

// Save 将收藏夹实体保存到数据库。
func (r *wishlistRepository) Save(ctx context.Context, wishlist *domain.Wishlist) error {
	model := toWishlistModel(wishlist)
	if model == nil {
		return nil
	}
	if err := r.db.WithContext(ctx).Save(model).Error; err != nil {
		return err
	}
	*wishlist = *toWishlist(model)
	return nil
}

// Delete 从数据库删除指定用户ID和收藏夹条目ID的记录。
func (r *wishlistRepository) Delete(ctx context.Context, userID, id uint64) error {
	return r.db.WithContext(ctx).Where("user_id = ? AND id = ?", userID, id).Delete(&WishlistModel{}).Error
}

// DeleteByProduct 从数据库删除指定用户ID和商品ID（SKUID）的记录。
func (r *wishlistRepository) DeleteByProduct(ctx context.Context, userID, skuID uint64) error {
	return r.db.WithContext(ctx).Where("user_id = ? AND sku_id = ?", userID, skuID).Delete(&WishlistModel{}).Error
}

// GetByID 获取指定用户ID和收藏夹条目ID的实体。
func (r *wishlistRepository) GetByID(ctx context.Context, userID, id uint64) (*domain.Wishlist, error) {
	var model WishlistModel
	if err := r.db.WithContext(ctx).Where("user_id = ? AND id = ?", userID, id).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toWishlist(&model), nil
}

// Get 获取指定用户ID和SKUID的收藏夹实体。
func (r *wishlistRepository) Get(ctx context.Context, userID, skuID uint64) (*domain.Wishlist, error) {
	var model WishlistModel
	if err := r.db.WithContext(ctx).Where("user_id = ? AND sku_id = ?", userID, skuID).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toWishlist(&model), nil
}

// List 从数据库列出指定用户ID的所有收藏夹记录，支持分页。
func (r *wishlistRepository) List(ctx context.Context, userID uint64, offset, limit int) ([]*domain.Wishlist, int64, error) {
	var list []*WishlistModel
	var total int64

	db := r.db.WithContext(ctx).Model(&WishlistModel{}).Where("user_id = ?", userID)
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := db.Offset(offset).Limit(limit).Order("id desc").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	results := make([]*domain.Wishlist, 0, len(list))
	for _, model := range list {
		results = append(results, toWishlist(model))
	}
	return results, total, nil
}

// Count 统计指定用户ID的收藏夹条目数量。
func (r *wishlistRepository) Count(ctx context.Context, userID uint64) (int64, error) {
	var total int64
	if err := r.db.WithContext(ctx).Model(&WishlistModel{}).Where("user_id = ?", userID).Count(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}

// Clear 清空指定用户的收藏夹。
func (r *wishlistRepository) Clear(ctx context.Context, userID uint64) error {
	return r.db.WithContext(ctx).Where("user_id = ?", userID).Delete(&WishlistModel{}).Error
}
