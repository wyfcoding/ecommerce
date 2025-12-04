package persistence

import (
	"context"
	"github.com/wyfcoding/ecommerce/internal/cart/domain/entity"     // 导入购物车模块的领域实体定义。
	"github.com/wyfcoding/ecommerce/internal/cart/domain/repository" // 导入购物车模块的领域仓储接口。

	"gorm.io/gorm" // 导入GORM ORM框架。
)

// cartRepository 是 CartRepository 接口的GORM实现。
// 它负责将购物车模块的领域实体映射到数据库，并执行持久化操作。
type cartRepository struct {
	db *gorm.DB // GORM数据库连接实例。
}

// NewCartRepository 创建并返回一个新的 cartRepository 实例。
// db: GORM数据库连接实例。
func NewCartRepository(db *gorm.DB) repository.CartRepository {
	return &cartRepository{db: db}
}

// Save 将购物车实体保存到数据库。
// 如果购物车已存在（通过ID），则更新其信息；如果不存在，则创建。
// 并且会级联保存购物车中的所有商品项。
func (r *cartRepository) Save(ctx context.Context, cart *entity.Cart) error {
	// GORM的Save方法会根据主键自动判断是创建还是更新。
	// Preload("Items") 和 Save 方法通常处理关联关系时需要特别注意，
	// 此处 Save(cart) 默认会处理 Cart 自身的更新，但对于 Items 的增删改，
	// 需要确保 GORM 配置了正确的级联操作，或在应用层手动处理 Items 的变化。
	return r.db.WithContext(ctx).Save(cart).Error
}

// GetByUserID 根据用户ID从数据库获取购物车记录，并预加载其关联的商品项。
// 如果找不到记录，则返回 nil 而不是错误，表示该用户还没有购物车。
func (r *cartRepository) GetByUserID(ctx context.Context, userID uint64) (*entity.Cart, error) {
	var cart entity.Cart
	// Preload("Items") 确保在获取购物车时，同时加载所有关联的商品项。
	if err := r.db.WithContext(ctx).Preload("Items").Where("user_id = ?", userID).First(&cart).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // 如果记录未找到，返回nil而不是错误。
		}
		return nil, err // 其他错误则返回。
	}
	return &cart, nil
}

// Delete 根据购物车ID从数据库删除购物车记录。
// 同时会删除购物车中所有关联的商品项（级联删除）。
func (r *cartRepository) Delete(ctx context.Context, id uint64) error {
	// Select("Items") 明确指示GORM在删除Cart时，级联删除其关联的Items。
	return r.db.WithContext(ctx).Select("Items").Delete(&entity.Cart{}, id).Error
}

// Clear 清空指定购物车ID的所有商品项，但不删除购物车本身。
func (r *cartRepository) Clear(ctx context.Context, cartID uint64) error {
	// 从数据库中删除所有与指定cartID关联的CartItem记录。
	return r.db.WithContext(ctx).Where("cart_id = ?", cartID).Delete(&entity.CartItem{}).Error
}
