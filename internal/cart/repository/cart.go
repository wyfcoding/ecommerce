package repository

import (
	"context"
	"fmt"
	"time"

	"ecommerce/internal/cart/model"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// CartRepo 定义了购物车数据的存储接口。
type CartRepo interface {
	// GetCartByUserID 根据用户ID获取购物车，包含所有购物车项。
	GetCartByUserID(ctx context.Context, userID uint64) (*model.Cart, error)
	// CreateCart 创建一个新的购物车。
	CreateCart(ctx context.Context, cart *model.Cart) (*model.Cart, error)
	// UpdateCart 更新购物车信息。
	UpdateCart(ctx context.Context, cart *model.Cart) (*model.Cart, error)
	// DeleteCart 逻辑删除购物车。
	DeleteCart(ctx context.Context, userID uint64) error
}

// CartItemRepo 定义了购物车项数据的存储接口。
type CartItemRepo interface {
	// GetCartItemByID 根据ID获取购物车项。
	GetCartItemByID(ctx context.Context, id uint64) (*model.CartItem, error)
	// GetCartItemBySKUID 根据用户ID和SKUID获取购物车项。
	GetCartItemBySKUID(ctx context.Context, userID, skuID uint64) (*model.CartItem, error)
	// CreateCartItem 创建一个新的购物车项。
	CreateCartItem(ctx context.Context, item *model.CartItem) (*model.CartItem, error)
	// UpdateCartItem 更新购物车项信息。
	UpdateCartItem(ctx context.Context, item *model.CartItem) (*model.CartItem, error)
	// DeleteCartItem 逻辑删除购物车项。
	DeleteCartItem(ctx context.Context, id uint64) error
	// DeleteCartItemsByUserID 逻辑删除某个用户的所有购物车项。
	DeleteCartItemsByUserID(ctx context.Context, userID uint64) error
	// DeleteCartItemsByIDs 批量逻辑删除指定ID的购物车项。
	DeleteCartItemsByIDs(ctx context.Context, userID uint64, itemIDs []uint64) error
	// ListCartItemsByUserID 根据用户ID获取所有购物车项。
	ListCartItemsByUserID(ctx context.Context, userID uint64) ([]*model.CartItem, error)
}

// cartRepoImpl 是 CartRepo 接口的 GORM 实现。
type cartRepoImpl struct {
	db *gorm.DB
}

// NewCartRepo 创建一个新的 CartRepo 实例。
func NewCartRepo(db *gorm.DB) CartRepo {
	return &cartRepoImpl{db: db}
}

// GetCartByUserID 实现 CartRepo.GetCartByUserID 方法。
func (r *cartRepoImpl) GetCartByUserID(ctx context.Context, userID uint64) (*model.Cart, error) {
	var cart model.Cart
	// 预加载所有购物车项
	if err := r.db.WithContext(ctx).Preload("Items").First(&cart, "user_id = ?", userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		zap.S().Errorf("failed to get cart by user id %d: %v", userID, err)
		return nil, fmt.Errorf("failed to get cart by user id: %w", err)
	}
	return &cart, nil
}

// CreateCart 实现 CartRepo.CreateCart 方法。
func (r *cartRepoImpl) CreateCart(ctx context.Context, cart *model.Cart) (*model.Cart, error) {
	if err := r.db.WithContext(ctx).Create(cart).Error; err != nil {
		zap.S().Errorf("failed to create cart for user %d: %v", cart.UserID, err)
		return nil, fmt.Errorf("failed to create cart: %w", err)
	}
	return cart, nil
}

// UpdateCart 实现 CartRepo.UpdateCart 方法。
func (r *cartRepoImpl) UpdateCart(ctx context.Context, cart *model.Cart) (*model.Cart, error) {
	if err := r.db.WithContext(ctx).Save(cart).Error; err != nil {
		zap.S().Errorf("failed to update cart for user %d: %v", cart.UserID, err)
		return nil, fmt.Errorf("failed to update cart: %w", err)
	}
	return cart, nil
}

// DeleteCart 实现 CartRepo.DeleteCart 方法 (逻辑删除)。
func (r *cartRepoImpl) DeleteCart(ctx context.Context, userID uint64) error {
	// 逻辑删除购物车主记录
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Delete(&model.Cart{}).Error; err != nil {
		zap.S().Errorf("failed to delete cart for user %d: %v", userID, err)
		return fmt.Errorf("failed to delete cart: %w", err)
	}
	// 同时逻辑删除所有关联的购物车项
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Delete(&model.CartItem{}).Error; err != nil {
		zap.S().Errorf("failed to delete cart items for user %d: %v", userID, err)
		return fmt.Errorf("failed to delete cart items: %w", err)
	}
	return nil
}

// cartItemRepoImpl 是 CartItemRepo 接口的 GORM 实现。
type cartItemRepoImpl struct {
	db *gorm.DB
}

// NewCartItemRepo 创建一个新的 CartItemRepo 实例。
func NewCartItemRepo(db *gorm.DB) CartItemRepo {
	return &cartItemRepoImpl{db: db}
}

// GetCartItemByID 实现 CartItemRepo.GetCartItemByID 方法。
func (r *cartItemRepoImpl) GetCartItemByID(ctx context.Context, id uint64) (*model.CartItem, error) {
	var item model.CartItem
	if err := r.db.WithContext(ctx).First(&item, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		zap.S().Errorf("failed to get cart item by id %d: %v", id, err)
		return nil, fmt.Errorf("failed to get cart item by id: %w", err)
	}
	return &item, nil
}

// GetCartItemBySKUID 实现 CartItemRepo.GetCartItemBySKUID 方法。
func (r *cartItemRepoImpl) GetCartItemBySKUID(ctx context.Context, userID, skuID uint64) (*model.CartItem, error) {
	var item model.CartItem
	if err := r.db.WithContext(ctx).Where("user_id = ? AND sku_id = ?", userID, skuID).First(&item).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		zap.S().Errorf("failed to get cart item by user %d and sku %d: %v", userID, skuID, err)
		return nil, fmt.Errorf("failed to get cart item by user id and sku id: %w", err)
	}
	return &item, nil
}

// CreateCartItem 实现 CartItemRepo.CreateCartItem 方法。
func (r *cartItemRepoImpl) CreateCartItem(ctx context.Context, item *model.CartItem) (*model.CartItem, error) {
	if err := r.db.WithContext(ctx).Create(item).Error; err != nil {
		zap.S().Errorf("failed to create cart item for user %d, sku %d: %v", item.UserID, item.SKUID, err)
		return nil, fmt.Errorf("failed to create cart item: %w", err)
	}
	return item, nil
}

// UpdateCartItem 实现 CartItemRepo.UpdateCartItem 方法。
func (r *cartItemRepoImpl) UpdateCartItem(ctx context.Context, item *model.CartItem) (*model.CartItem, error) {
	if err := r.db.WithContext(ctx).Save(item).Error; err != nil {
		zap.S().Errorf("failed to update cart item %d for user %d: %v", item.ID, item.UserID, err)
		return nil, fmt.Errorf("failed to update cart item: %w", err)
	}
	return item, nil
}

// DeleteCartItem 实现 CartItemRepo.DeleteCartItem 方法 (逻辑删除)。
func (r *cartItemRepoImpl) DeleteCartItem(ctx context.Context, id uint64) error {
	if err := r.db.WithContext(ctx).Delete(&model.CartItem{}, id).Error; err != nil {
		zap.S().Errorf("failed to delete cart item %d: %v", id, err)
		return fmt.Errorf("failed to delete cart item: %w", err)
	}
	return nil
}

// DeleteCartItemsByUserID 实现 CartItemRepo.DeleteCartItemsByUserID 方法 (逻辑删除)。
func (r *cartItemRepoImpl) DeleteCartItemsByUserID(ctx context.Context, userID uint64) error {
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Delete(&model.CartItem{}).Error; err != nil {
		zap.S().Errorf("failed to delete cart items for user %d: %v", userID, err)
		return fmt.Errorf("failed to delete cart items by user id: %w", err)
	}
	return nil
}

// DeleteCartItemsByIDs 实现 CartItemRepo.DeleteCartItemsByIDs 方法 (批量逻辑删除)。
func (r *cartItemRepoImpl) DeleteCartItemsByIDs(ctx context.Context, userID uint64, itemIDs []uint64) error {
	if err := r.db.WithContext(ctx).Where("user_id = ? AND id IN (?) ", userID, itemIDs).Delete(&model.CartItem{}).Error; err != nil {
		zap.S().Errorf("failed to delete cart items %v for user %d: %v", itemIDs, userID, err)
		return fmt.Errorf("failed to delete cart items by ids: %w", err)
	}
	return nil
}

// ListCartItemsByUserID 实现 CartItemRepo.ListCartItemsByUserID 方法。
func (r *cartItemRepoImpl) ListCartItemsByUserID(ctx context.Context, userID uint64) ([]*model.CartItem, error) {
	var items []*model.CartItem
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&items).Error; err != nil {
		zap.S().Errorf("failed to list cart items for user %d: %v", userID, err)
		return nil, fmt.Errorf("failed to list cart items by user id: %w", err)
	}
	return items, nil
}
