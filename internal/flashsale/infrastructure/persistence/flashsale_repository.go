package persistence

import (
	"context"
	"errors"

	"github.com/wyfcoding/ecommerce/internal/flashsale/domain"

	"gorm.io/gorm"
)

type flashSaleRepository struct {
	db *gorm.DB
}

// NewFlashSaleRepository 创建并返回一个新的 flashSaleRepository 实例。
func NewFlashSaleRepository(db *gorm.DB) domain.FlashSaleRepository {
	return &flashSaleRepository{db: db}
}

// --- 活动管理 (Flashsale methods) ---

// SaveFlashsale 将秒杀活动实体保存到数据库。
func (r *flashSaleRepository) SaveFlashsale(ctx context.Context, flashsale *domain.Flashsale) error {
	return r.db.WithContext(ctx).Save(flashsale).Error
}

// GetFlashsale 根据ID从数据库获取秒杀活动记录。
func (r *flashSaleRepository) GetFlashsale(ctx context.Context, id uint64) (*domain.Flashsale, error) {
	var flashsale domain.Flashsale
	if err := r.db.WithContext(ctx).First(&flashsale, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrFlashsaleNotFound
		}
		return nil, err
	}
	return &flashsale, nil
}

// ListFlashsales 从数据库列出所有秒杀活动记录。
func (r *flashSaleRepository) ListFlashsales(ctx context.Context, status *domain.FlashsaleStatus, offset, limit int) ([]*domain.Flashsale, int64, error) {
	var list []*domain.Flashsale
	var total int64

	db := r.db.WithContext(ctx).Model(&domain.Flashsale{})
	if status != nil {
		db = db.Where("status = ?", *status)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := db.Offset(offset).Limit(limit).Order("start_time asc").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// UpdateStock 更新秒杀活动的商品库存。
func (r *flashSaleRepository) UpdateStock(ctx context.Context, id uint64, quantity int32) error {
	return r.db.WithContext(ctx).Model(&domain.Flashsale{}).
		Where("id = ? AND sold_count + ? <= total_stock", id, quantity).
		UpdateColumn("sold_count", gorm.Expr("sold_count + ?", quantity)).Error
}

// --- 订单管理 (FlashsaleOrder methods) ---

// SaveOrder 将秒杀订单实体保存到数据库。
func (r *flashSaleRepository) SaveOrder(ctx context.Context, order *domain.FlashsaleOrder) error {
	return r.db.WithContext(ctx).Save(order).Error
}

// GetOrder 根据ID从数据库获取秒杀订单记录。
func (r *flashSaleRepository) GetOrder(ctx context.Context, id uint64) (*domain.FlashsaleOrder, error) {
	var order domain.FlashsaleOrder
	if err := r.db.WithContext(ctx).First(&order, id).Error; err != nil {
		return nil, err
	}
	return &order, nil
}

// GetUserOrders 获取指定用户在某个秒杀活动中的所有订单记录。
func (r *flashSaleRepository) GetUserOrders(ctx context.Context, userID, flashsaleID uint64) ([]*domain.FlashsaleOrder, error) {
	var list []*domain.FlashsaleOrder
	if err := r.db.WithContext(ctx).Where("user_id = ? AND flashsale_id = ?", userID, flashsaleID).Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

// CountUserBought 统计指定用户在某个秒杀活动中已购买的商品数量。
func (r *flashSaleRepository) CountUserBought(ctx context.Context, userID, flashsaleID uint64) (int32, error) {
	var total int64
	if err := r.db.WithContext(ctx).Model(&domain.FlashsaleOrder{}).
		Where("user_id = ? AND flashsale_id = ? AND status != ?", userID, flashsaleID, domain.FlashsaleOrderStatusCancelled).
		Select("COALESCE(SUM(quantity), 0)").
		Scan(&total).Error; err != nil {
		return 0, err
	}
	return int32(total), nil
}
