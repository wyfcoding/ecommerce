package persistence

import (
	"context"
	"errors"
	"github.com/wyfcoding/ecommerce/internal/flashsale/domain/entity"
	"github.com/wyfcoding/ecommerce/internal/flashsale/domain/repository"

	"gorm.io/gorm"
)

type flashSaleRepository struct {
	db *gorm.DB
}

func NewFlashSaleRepository(db *gorm.DB) repository.FlashSaleRepository {
	return &flashSaleRepository{db: db}
}

// 活动管理
func (r *flashSaleRepository) SaveFlashsale(ctx context.Context, flashsale *entity.Flashsale) error {
	return r.db.WithContext(ctx).Save(flashsale).Error
}

func (r *flashSaleRepository) GetFlashsale(ctx context.Context, id uint64) (*entity.Flashsale, error) {
	var flashsale entity.Flashsale
	if err := r.db.WithContext(ctx).First(&flashsale, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, entity.ErrFlashsaleNotFound
		}
		return nil, err
	}
	return &flashsale, nil
}

func (r *flashSaleRepository) ListFlashsales(ctx context.Context, status *entity.FlashsaleStatus, offset, limit int) ([]*entity.Flashsale, int64, error) {
	var list []*entity.Flashsale
	var total int64

	db := r.db.WithContext(ctx).Model(&entity.Flashsale{})
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

func (r *flashSaleRepository) UpdateStock(ctx context.Context, id uint64, quantity int32) error {
	return r.db.WithContext(ctx).Model(&entity.Flashsale{}).
		Where("id = ? AND sold_count + ? <= total_stock", id, quantity).
		UpdateColumn("sold_count", gorm.Expr("sold_count + ?", quantity)).Error
}

// 订单管理
func (r *flashSaleRepository) SaveOrder(ctx context.Context, order *entity.FlashsaleOrder) error {
	return r.db.WithContext(ctx).Save(order).Error
}

func (r *flashSaleRepository) GetOrder(ctx context.Context, id uint64) (*entity.FlashsaleOrder, error) {
	var order entity.FlashsaleOrder
	if err := r.db.WithContext(ctx).First(&order, id).Error; err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *flashSaleRepository) GetUserOrders(ctx context.Context, userID, flashsaleID uint64) ([]*entity.FlashsaleOrder, error) {
	var list []*entity.FlashsaleOrder
	if err := r.db.WithContext(ctx).Where("user_id = ? AND flashsale_id = ?", userID, flashsaleID).Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *flashSaleRepository) CountUserBought(ctx context.Context, userID, flashsaleID uint64) (int32, error) {
	var total int64
	if err := r.db.WithContext(ctx).Model(&entity.FlashsaleOrder{}).
		Where("user_id = ? AND flashsale_id = ? AND status != ?", userID, flashsaleID, entity.FlashsaleOrderStatusCancelled).
		Select("COALESCE(SUM(quantity), 0)").
		Scan(&total).Error; err != nil {
		return 0, err
	}
	return int32(total), nil
}
