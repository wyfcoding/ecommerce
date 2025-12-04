package persistence

import (
	"context"
	"errors"
	"github.com/wyfcoding/ecommerce/internal/flashsale/domain/entity"     // 导入秒杀模块的领域实体定义。
	"github.com/wyfcoding/ecommerce/internal/flashsale/domain/repository" // 导入秒杀模块的领域仓储接口。

	"gorm.io/gorm" // 导入GORM ORM框架。
)

type flashSaleRepository struct {
	db *gorm.DB // GORM数据库连接实例。
}

// NewFlashSaleRepository 创建并返回一个新的 flashSaleRepository 实例。
// db: GORM数据库连接实例。
func NewFlashSaleRepository(db *gorm.DB) repository.FlashSaleRepository {
	return &flashSaleRepository{db: db}
}

// --- 活动管理 (Flashsale methods) ---

// SaveFlashsale 将秒杀活动实体保存到数据库。
// 如果活动已存在（通过ID），则更新其信息；如果不存在，则创建。
func (r *flashSaleRepository) SaveFlashsale(ctx context.Context, flashsale *entity.Flashsale) error {
	return r.db.WithContext(ctx).Save(flashsale).Error
}

// GetFlashsale 根据ID从数据库获取秒杀活动记录。
// 如果记录未找到，则返回 entity.ErrFlashsaleNotFound 错误。
func (r *flashSaleRepository) GetFlashsale(ctx context.Context, id uint64) (*entity.Flashsale, error) {
	var flashsale entity.Flashsale
	if err := r.db.WithContext(ctx).First(&flashsale, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, entity.ErrFlashsaleNotFound // 返回自定义的“未找到”错误。
		}
		return nil, err
	}
	return &flashsale, nil
}

// ListFlashsales 从数据库列出所有秒杀活动记录，支持通过状态过滤和分页。
func (r *flashSaleRepository) ListFlashsales(ctx context.Context, status *entity.FlashsaleStatus, offset, limit int) ([]*entity.Flashsale, int64, error) {
	var list []*entity.Flashsale
	var total int64

	db := r.db.WithContext(ctx).Model(&entity.Flashsale{})
	if status != nil { // 如果提供了状态，则按状态过滤。
		db = db.Where("status = ?", *status)
	}

	// 统计总记录数。
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 应用分页和排序。
	if err := db.Offset(offset).Limit(limit).Order("start_time asc").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// UpdateStock 更新秒杀活动的商品库存。
// 这是一个原子性操作，通过数据库层的乐观锁（使用WHERE子句检查条件）来防止超卖。
// id: 秒杀活动ID。
// quantity: 要减少的库存数量。
func (r *flashSaleRepository) UpdateStock(ctx context.Context, id uint64, quantity int32) error {
	// GORM的UpdateColumn方法直接更新列，并配合Where子句实现乐观锁。
	// 只有当 sold_count + quantity 不会超过 total_stock 时才执行更新。
	return r.db.WithContext(ctx).Model(&entity.Flashsale{}).
		Where("id = ? AND sold_count + ? <= total_stock", id, quantity).
		UpdateColumn("sold_count", gorm.Expr("sold_count + ?", quantity)).Error
}

// --- 订单管理 (FlashsaleOrder methods) ---

// SaveOrder 将秒杀订单实体保存到数据库。
func (r *flashSaleRepository) SaveOrder(ctx context.Context, order *entity.FlashsaleOrder) error {
	return r.db.WithContext(ctx).Save(order).Error
}

// GetOrder 根据ID从数据库获取秒杀订单记录。
func (r *flashSaleRepository) GetOrder(ctx context.Context, id uint64) (*entity.FlashsaleOrder, error) {
	var order entity.FlashsaleOrder
	if err := r.db.WithContext(ctx).First(&order, id).Error; err != nil {
		return nil, err
	}
	return &order, nil
}

// GetUserOrders 获取指定用户在某个秒杀活动中的所有订单记录。
func (r *flashSaleRepository) GetUserOrders(ctx context.Context, userID, flashsaleID uint64) ([]*entity.FlashsaleOrder, error) {
	var list []*entity.FlashsaleOrder
	if err := r.db.WithContext(ctx).Where("user_id = ? AND flashsale_id = ?", userID, flashsaleID).Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

// CountUserBought 统计指定用户在某个秒杀活动中已购买的商品数量。
// 只统计未取消的订单。
func (r *flashSaleRepository) CountUserBought(ctx context.Context, userID, flashsaleID uint64) (int32, error) {
	var total int64
	if err := r.db.WithContext(ctx).Model(&entity.FlashsaleOrder{}).
		Where("user_id = ? AND flashsale_id = ? AND status != ?", userID, flashsaleID, entity.FlashsaleOrderStatusCancelled).
		Select("COALESCE(SUM(quantity), 0)"). // 使用COALESCE处理SUM结果可能为NULL的情况。
		Scan(&total).Error; err != nil {
		return 0, err
	}
	return int32(total), nil
}
