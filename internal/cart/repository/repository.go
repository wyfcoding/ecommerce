package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"

	"ecommerce/internal/cart/model"
)

// CartRepository 定义了购物车数据仓库的接口
// 这样做可以实现依赖反转，方便测试时使用模拟对象
type CartRepository interface {
	GetCartByUserID(ctx context.Context, userID uint) (*model.Cart, error)
	CreateCart(ctx context.Context, cart *model.Cart) error
	UpdateCart(ctx context.Context, cart *model.Cart) error
	AddItemToCart(ctx context.Context, item *model.CartItem) error
	UpdateCartItem(ctx context.Context, item *model.CartItem) error
	DeleteCartItem(ctx context.Context, itemID uint) error
	GetCartItem(ctx context.Context, cartID, productID uint) (*model.CartItem, error)
}

// cartRepository 是 CartRepository 的具体实现
type cartRepository struct {
	db    *gorm.DB
	cache *redis.Client
}

// NewCartRepository 创建一个新的 cartRepository 实例
// 参数 db 是 GORM 数据库连接实例
// 参数 cache 是 Redis 客户端实例
func NewCartRepository(db *gorm.DB, cache *redis.Client) CartRepository {
	return &cartRepository{db: db, cache: cache}
}

// getCartCacheKey 生成用户购物车的 Redis 缓存键
func (r *cartRepository) getCartCacheKey(userID uint) string {
	return fmt.Sprintf("cart:%d", userID)
}

// GetCartByUserID 根据用户ID获取购物车
// 它首先尝试从 Redis 缓存中获取，如果缓存未命中，则查询数据库并将结果存入缓存
func (r *cartRepository) GetCartByUserID(ctx context.Context, userID uint) (*model.Cart, error) {
	// 1. 尝试从缓存获取
	cacheKey := r.getCartCacheKey(userID)
	val, err := r.cache.Get(ctx, cacheKey).Result()
	if err == nil {
		var cart model.Cart
		if json.Unmarshal([]byte(val), &cart) == nil {
			// 缓存命中，直接返回
			return &cart, nil
		}
	}

	// 2. 缓存未命中或解析失败，查询数据库
	var cart model.Cart
	// Preload("Items") 会同时加载购物车中的所有商品项
	if err := r.db.WithContext(ctx).Preload("Items").First(&cart, "user_id = ?", userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// 如果记录不存在，返回 nil, nil 表示正常但没有购物车
			return nil, nil
		}
		return nil, fmt.Errorf("数据库查询失败: %w", err)
	}

	// 3. 将从数据库查询到的结果存入缓存
	cartJSON, err := json.Marshal(&cart)
	if err == nil {
		// 设置缓存，过期时间为1小时
		r.cache.Set(ctx, cacheKey, cartJSON, time.Hour).Err()
	}

	return &cart, nil
}

// CreateCart 创建一个新的购物车
// 创建成功后，会清除对应的缓存，以确保数据一致性
func (r *cartRepository) CreateCart(ctx context.Context, cart *model.Cart) error {
	if err := r.db.WithContext(ctx).Create(cart).Error; err != nil {
		return fmt.Errorf("创建购物车失败: %w", err)
	}
	// 使缓存失效
	r.cache.Del(ctx, r.getCartCacheKey(cart.UserID))
	return nil
}

// UpdateCart 更新购物车信息 (主要用于更新整个购物车，例如合并操作)
// 操作成功后，会清除对应的缓存
func (r *cartRepository) UpdateCart(ctx context.Context, cart *model.Cart) error {
	if err := r.db.WithContext(ctx).Save(cart).Error; err != nil {
		return fmt.Errorf("更新购物车失败: %w", err)
	}
	// 使缓存失效
	r.cache.Del(ctx, r.getCartCacheKey(cart.UserID))
	return nil
}

// AddItemToCart 向购物车添加一个新商品项
// 操作成功后，会清除对应的缓存
func (r *cartRepository) AddItemToCart(ctx context.Context, item *model.CartItem) error {
	if err := r.db.WithContext(ctx).Create(item).Error; err != nil {
		return fmt.Errorf("添加商品项失败: %w", err)
	}
	// 查询 cart 以获取 userID
	var cart model.Cart
	r.db.WithContext(ctx).First(&cart, item.CartID)
	// 使缓存失效
	r.cache.Del(ctx, r.getCartCacheKey(cart.UserID))
	return nil
}

// UpdateCartItem 更新购物车中某个商品项的信息（如数量）
// 操作成功后，会清除对应的缓存
func (r *cartRepository) UpdateCartItem(ctx context.Context, item *model.CartItem) error {
	if err := r.db.WithContext(ctx).Save(item).Error; err != nil {
		return fmt.Errorf("更新商品项失败: %w", err)
	}
	var cart model.Cart
	r.db.WithContext(ctx).First(&cart, item.CartID)
	r.cache.Del(ctx, r.getCartCacheKey(cart.UserID))
	return nil
}

// DeleteCartItem 从购物车中删除一个商品项
// 操作成功后，会清除对应的缓存
func (r *cartRepository) DeleteCartItem(ctx context.Context, itemID uint) error {
    // 在删除前，需要先获取 item 信息以找到 userID
    var item model.CartItem
    if err := r.db.WithContext(ctx).First(&item, itemID).Error; err != nil {
        return fmt.Errorf("删除前查找商品项失败: %w", err)
    }
    var cart model.Cart
    if err := r.db.WithContext(ctx).First(&cart, item.CartID).Error; err != nil {
        return fmt.Errorf("删除前查找购物车失败: %w", err)
    }

	if err := r.db.WithContext(ctx).Delete(&model.CartItem{}, itemID).Error; err != nil {
		return fmt.Errorf("删除商品项失败: %w", err)
	}

	r.cache.Del(ctx, r.getCartCacheKey(cart.UserID))
	return nil
}

// GetCartItem 获取购物车中的特定商品项
// 用于判断某个商品是否已存在于购物车中
func (r *cartRepository) GetCartItem(ctx context.Context, cartID, productID uint) (*model.CartItem, error) {
	var item model.CartItem
	if err := r.db.WithContext(ctx).Where("cart_id = ? AND product_id = ?", cartID, productID).First(&item).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // 商品不存在于购物车中
		}
		return nil, fmt.Errorf("查询商品项失败: %w", err)
	}
	return &item, nil
}