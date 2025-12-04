package persistence

import (
	"context"
	"errors" // 导入标准错误处理库。

	"github.com/wyfcoding/ecommerce/internal/wishlist/domain/entity"     // 导入收藏夹领域的实体定义。
	"github.com/wyfcoding/ecommerce/internal/wishlist/domain/repository" // 导入收藏夹领域的仓储接口。

	"gorm.io/gorm" // 导入GORM ORM框架。
)

type wishlistRepository struct {
	db *gorm.DB // GORM数据库连接实例。
}

// NewWishlistRepository 创建并返回一个新的 wishlistRepository 实例。
func NewWishlistRepository(db *gorm.DB) repository.WishlistRepository {
	return &wishlistRepository{db: db}
}

// Save 将收藏夹实体保存到数据库。
// 如果实体已存在，则更新；如果不存在，则创建。
func (r *wishlistRepository) Save(ctx context.Context, wishlist *entity.Wishlist) error {
	return r.db.WithContext(ctx).Save(wishlist).Error
}

// Delete 从数据库删除指定用户ID和收藏夹条目ID的记录。
// 确保用户只能删除自己的收藏夹条目。
func (r *wishlistRepository) Delete(ctx context.Context, userID, id uint64) error {
	return r.db.WithContext(ctx).Where("user_id = ? AND id = ?", userID, id).Delete(&entity.Wishlist{}).Error
}

// Get 获取指定用户ID和SKUID的收藏夹实体。
// 如果记录未找到，则返回nil。
func (r *wishlistRepository) Get(ctx context.Context, userID, skuID uint64) (*entity.Wishlist, error) {
	var wishlist entity.Wishlist
	// 按用户ID和SKUID过滤，确保获取特定用户的特定商品收藏。
	if err := r.db.WithContext(ctx).Where("user_id = ? AND sku_id = ?", userID, skuID).First(&wishlist).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // 如果记录未找到，返回nil。
		}
		return nil, err // 其他错误则返回。
	}
	return &wishlist, nil
}

// List 从数据库列出指定用户ID的所有收藏夹记录，支持分页。
func (r *wishlistRepository) List(ctx context.Context, userID uint64, offset, limit int) ([]*entity.Wishlist, int64, error) {
	var list []*entity.Wishlist
	var total int64

	db := r.db.WithContext(ctx).Model(&entity.Wishlist{}).Where("user_id = ?", userID)
	if err := db.Count(&total).Error; err != nil { // 统计总记录数。
		return nil, 0, err
	}

	// 应用分页和排序（按ID降序）。
	if err := db.Offset(offset).Limit(limit).Order("id desc").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// Count 统计指定用户ID的收藏夹条目数量。
func (r *wishlistRepository) Count(ctx context.Context, userID uint64) (int64, error) {
	var total int64
	// 按用户ID过滤并统计数量。
	if err := r.db.WithContext(ctx).Model(&entity.Wishlist{}).Where("user_id = ?", userID).Count(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}
