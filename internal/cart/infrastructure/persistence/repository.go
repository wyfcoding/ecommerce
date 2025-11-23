package persistence

import (
	"context"
	"ecommerce/internal/cart/domain/entity"
	"ecommerce/internal/cart/domain/repository"

	"gorm.io/gorm"
)

type cartRepository struct {
	db *gorm.DB
}

func NewCartRepository(db *gorm.DB) repository.CartRepository {
	return &cartRepository{db: db}
}

func (r *cartRepository) Save(ctx context.Context, cart *entity.Cart) error {
	return r.db.WithContext(ctx).Save(cart).Error
}

func (r *cartRepository) GetByUserID(ctx context.Context, userID uint64) (*entity.Cart, error) {
	var cart entity.Cart
	if err := r.db.WithContext(ctx).Preload("Items").Where("user_id = ?", userID).First(&cart).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &cart, nil
}

func (r *cartRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Select("Items").Delete(&entity.Cart{}, id).Error
}

func (r *cartRepository) Clear(ctx context.Context, cartID uint64) error {
	return r.db.WithContext(ctx).Where("cart_id = ?", cartID).Delete(&entity.CartItem{}).Error
}
