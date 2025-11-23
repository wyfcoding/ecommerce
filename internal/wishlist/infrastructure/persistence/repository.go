package persistence

import (
	"context"
	"ecommerce/internal/wishlist/domain/entity"
	"ecommerce/internal/wishlist/domain/repository"
	"errors"

	"gorm.io/gorm"
)

type wishlistRepository struct {
	db *gorm.DB
}

func NewWishlistRepository(db *gorm.DB) repository.WishlistRepository {
	return &wishlistRepository{db: db}
}

func (r *wishlistRepository) Save(ctx context.Context, wishlist *entity.Wishlist) error {
	return r.db.WithContext(ctx).Save(wishlist).Error
}

func (r *wishlistRepository) Delete(ctx context.Context, userID, id uint64) error {
	return r.db.WithContext(ctx).Where("user_id = ? AND id = ?", userID, id).Delete(&entity.Wishlist{}).Error
}

func (r *wishlistRepository) Get(ctx context.Context, userID, skuID uint64) (*entity.Wishlist, error) {
	var wishlist entity.Wishlist
	if err := r.db.WithContext(ctx).Where("user_id = ? AND sku_id = ?", userID, skuID).First(&wishlist).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &wishlist, nil
}

func (r *wishlistRepository) List(ctx context.Context, userID uint64, offset, limit int) ([]*entity.Wishlist, int64, error) {
	var list []*entity.Wishlist
	var total int64

	db := r.db.WithContext(ctx).Model(&entity.Wishlist{}).Where("user_id = ?", userID)
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := db.Offset(offset).Limit(limit).Order("id desc").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

func (r *wishlistRepository) Count(ctx context.Context, userID uint64) (int64, error) {
	var total int64
	if err := r.db.WithContext(ctx).Model(&entity.Wishlist{}).Where("user_id = ?", userID).Count(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}
