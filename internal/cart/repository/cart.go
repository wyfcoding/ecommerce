package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"ecommerce/internal/cart/model"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

type CartRepo interface {
	CreateCart(ctx context.Context, cart *model.Cart) error
	GetCartByID(ctx context.Context, id uint64) (*model.Cart, error)
	GetCartByUserID(ctx context.Context, userID uint64) (*model.Cart, error)
	UpdateCart(ctx context.Context, cart *model.Cart) error
	DeleteCart(ctx context.Context, id uint64) error
}

type CartItemRepo interface {
	CreateCartItem(ctx context.Context, item *model.CartItem) error
	GetCartItemByID(ctx context.Context, id uint64) (*model.CartItem, error)
	GetCartItemBySkuID(ctx context.Context, cartID, skuID uint64) (*model.CartItem, error)
	GetCartItemsByCartID(ctx context.Context, cartID uint64) ([]*model.CartItem, error)
	GetSelectedCartItems(ctx context.Context, cartID uint64) ([]*model.CartItem, error)
	UpdateCartItem(ctx context.Context, item *model.CartItem) error
	DeleteCartItem(ctx context.Context, id uint64) error
	DeleteCartItemsByCartID(ctx context.Context, cartID uint64) error
}

type cartRepository struct {
	db    *gorm.DB
	redis *redis.Client
}

func NewCartRepository(db *gorm.DB, redis *redis.Client) CartRepo {
	return &cartRepository{db: db, redis: redis}
}

func (r *cartRepository) CreateCart(ctx context.Context, cart *model.Cart) error {
	if err := r.db.WithContext(ctx).Create(cart).Error; err != nil {
		return err
	}
	// 缓存到Redis
	r.cacheCart(ctx, cart)
	return nil
}

func (r *cartRepository) GetCartByID(ctx context.Context, id uint64) (*model.Cart, error) {
	// 先从缓存获取
	if cart := r.getCartFromCache(ctx, fmt.Sprintf("cart:id:%d", id)); cart != nil {
		return cart, nil
	}

	var cart model.Cart
	if err := r.db.WithContext(ctx).First(&cart, id).Error; err != nil {
		return nil, err
	}

	r.cacheCart(ctx, &cart)
	return &cart, nil
}

func (r *cartRepository) GetCartByUserID(ctx context.Context, userID uint64) (*model.Cart, error) {
	// 先从缓存获取
	cacheKey := fmt.Sprintf("cart:user:%d", userID)
	if cart := r.getCartFromCache(ctx, cacheKey); cart != nil {
		return cart, nil
	}

	var cart model.Cart
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&cart).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}

	r.cacheCart(ctx, &cart)
	return &cart, nil
}

func (r *cartRepository) UpdateCart(ctx context.Context, cart *model.Cart) error {
	if err := r.db.WithContext(ctx).Save(cart).Error; err != nil {
		return err
	}
	// 更新缓存
	r.cacheCart(ctx, cart)
	return nil
}

func (r *cartRepository) DeleteCart(ctx context.Context, id uint64) error {
	if err := r.db.WithContext(ctx).Delete(&model.Cart{}, id).Error; err != nil {
		return err
	}
	// 删除缓存
	r.redis.Del(ctx, fmt.Sprintf("cart:id:%d", id))
	return nil
}

func (r *cartRepository) cacheCart(ctx context.Context, cart *model.Cart) {
	if r.redis == nil {
		return
	}
	data, _ := json.Marshal(cart)
	r.redis.Set(ctx, fmt.Sprintf("cart:id:%d", cart.UserID), data, 30*time.Minute)
	r.redis.Set(ctx, fmt.Sprintf("cart:user:%d", cart.UserID), data, 30*time.Minute)
}

func (r *cartRepository) getCartFromCache(ctx context.Context, key string) *model.Cart {
	if r.redis == nil {
		return nil
	}
	data, err := r.redis.Get(ctx, key).Bytes()
	if err != nil {
		return nil
	}
	var cart model.Cart
	if err := json.Unmarshal(data, &cart); err != nil {
		return nil
	}
	return &cart
}

type cartItemRepository struct {
	db    *gorm.DB
	redis *redis.Client
}

func NewCartItemRepository(db *gorm.DB, redis *redis.Client) CartItemRepo {
	return &cartItemRepository{db: db, redis: redis}
}

func (r *cartItemRepository) CreateCartItem(ctx context.Context, item *model.CartItem) error {
	if err := r.db.WithContext(ctx).Create(item).Error; err != nil {
		return err
	}
	r.invalidateCartCache(ctx, item.UserID)
	return nil
}

func (r *cartItemRepository) GetCartItemByID(ctx context.Context, id uint64) (*model.CartItem, error) {
	var item model.CartItem
	if err := r.db.WithContext(ctx).First(&item, id).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *cartItemRepository) GetCartItemBySkuID(ctx context.Context, cartID, skuID uint64) (*model.CartItem, error) {
	var item model.CartItem
	if err := r.db.WithContext(ctx).Where("user_id = ? AND sku_id = ?", cartID, skuID).First(&item).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &item, nil
}

func (r *cartItemRepository) GetCartItemsByCartID(ctx context.Context, cartID uint64) ([]*model.CartItem, error) {
	// 先从缓存获取
	cacheKey := fmt.Sprintf("cart:items:%d", cartID)
	if items := r.getCartItemsFromCache(ctx, cacheKey); items != nil {
		return items, nil
	}

	var items []*model.CartItem
	if err := r.db.WithContext(ctx).Where("user_id = ?", cartID).Find(&items).Error; err != nil {
		return nil, err
	}

	r.cacheCartItems(ctx, cacheKey, items)
	return items, nil
}

func (r *cartItemRepository) GetSelectedCartItems(ctx context.Context, cartID uint64) ([]*model.CartItem, error) {
	var items []*model.CartItem
	if err := r.db.WithContext(ctx).Where("user_id = ? AND selected = ?", cartID, true).Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (r *cartItemRepository) UpdateCartItem(ctx context.Context, item *model.CartItem) error {
	if err := r.db.WithContext(ctx).Save(item).Error; err != nil {
		return err
	}
	r.invalidateCartCache(ctx, item.UserID)
	return nil
}

func (r *cartItemRepository) DeleteCartItem(ctx context.Context, id uint64) error {
	var item model.CartItem
	if err := r.db.WithContext(ctx).First(&item, id).Error; err != nil {
		return err
	}
	if err := r.db.WithContext(ctx).Delete(&model.CartItem{}, id).Error; err != nil {
		return err
	}
	r.invalidateCartCache(ctx, item.UserID)
	return nil
}

func (r *cartItemRepository) DeleteCartItemsByCartID(ctx context.Context, cartID uint64) error {
	if err := r.db.WithContext(ctx).Where("user_id = ?", cartID).Delete(&model.CartItem{}).Error; err != nil {
		return err
	}
	r.invalidateCartCache(ctx, cartID)
	return nil
}

func (r *cartItemRepository) invalidateCartCache(ctx context.Context, cartID uint64) {
	if r.redis == nil {
		return
	}
	r.redis.Del(ctx, fmt.Sprintf("cart:items:%d", cartID))
}

func (r *cartItemRepository) cacheCartItems(ctx context.Context, key string, items []*model.CartItem) {
	if r.redis == nil {
		return
	}
	data, _ := json.Marshal(items)
	r.redis.Set(ctx, key, data, 30*time.Minute)
}

func (r *cartItemRepository) getCartItemsFromCache(ctx context.Context, key string) []*model.CartItem {
	if r.redis == nil {
		return nil
	}
	data, err := r.redis.Get(ctx, key).Bytes()
	if err != nil {
		return nil
	}
	var items []*model.CartItem
	if err := json.Unmarshal(data, &items); err != nil {
		return nil
	}
	return items
}
