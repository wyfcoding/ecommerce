package persistence

import (
	"context"
	"ecommerce/internal/order/domain/entity"
	"ecommerce/internal/order/domain/repository"
	"errors"

	"gorm.io/gorm"
)

type orderRepository struct {
	db *gorm.DB
}

func NewOrderRepository(db *gorm.DB) repository.OrderRepository {
	return &orderRepository{db: db}
}

func (r *orderRepository) Save(ctx context.Context, order *entity.Order) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(order).Error; err != nil {
			return err
		}
		// Save items
		for _, item := range order.Items {
			if item.ID == 0 {
				item.OrderID = uint64(order.ID)
			}
			if err := tx.Save(item).Error; err != nil {
				return err
			}
		}
		// Save logs
		for _, log := range order.Logs {
			if log.ID == 0 {
				log.OrderID = uint64(order.ID)
			}
			if err := tx.Save(log).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *orderRepository) GetByID(ctx context.Context, id uint64) (*entity.Order, error) {
	var order entity.Order
	if err := r.db.WithContext(ctx).Preload("Items").Preload("Logs").First(&order, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &order, nil
}

func (r *orderRepository) GetByOrderNo(ctx context.Context, orderNo string) (*entity.Order, error) {
	var order entity.Order
	if err := r.db.WithContext(ctx).Preload("Items").Preload("Logs").Where("order_no = ?", orderNo).First(&order).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &order, nil
}

func (r *orderRepository) List(ctx context.Context, userID uint64, status *entity.OrderStatus, offset, limit int) ([]*entity.Order, int64, error) {
	var list []*entity.Order
	var total int64

	db := r.db.WithContext(ctx).Model(&entity.Order{})

	if userID > 0 {
		db = db.Where("user_id = ?", userID)
	}
	if status != nil {
		db = db.Where("status = ?", *status)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := db.Preload("Items").Offset(offset).Limit(limit).Order("created_at desc").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}
