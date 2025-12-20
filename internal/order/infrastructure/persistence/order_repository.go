package persistence

import (
	"context"
	"errors"

	"github.com/wyfcoding/ecommerce/internal/order/domain"
	"github.com/wyfcoding/pkg/databases/sharding"

	"gorm.io/gorm"
)

type orderRepository struct {
	sharding *sharding.Manager
}

func NewOrderRepository(sharding *sharding.Manager) domain.OrderRepository {
	return &orderRepository{sharding: sharding}
}

func (r *orderRepository) Save(ctx context.Context, order *domain.Order) error {
	db := r.sharding.GetDB(order.UserID)
	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(order).Error; err != nil {
			return err
		}
		for _, item := range order.Items {
			if item.ID == 0 {
				item.OrderID = uint64(order.ID)
			}
			if err := tx.Save(item).Error; err != nil {
				return err
			}
		}
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

func (r *orderRepository) FindByID(ctx context.Context, id uint) (*domain.Order, error) {
	// TODO: Support sharding by ID or UserID hint
	db := r.sharding.GetDB(0)
	var order domain.Order
	if err := db.WithContext(ctx).Preload("Items").Preload("Logs").First(&order, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &order, nil
}

func (r *orderRepository) FindByOrderNo(ctx context.Context, orderNo string) (*domain.Order, error) {
	db := r.sharding.GetDB(0)
	var order domain.Order
	if err := db.WithContext(ctx).Preload("Items").Preload("Logs").Where("order_no = ?", orderNo).First(&order).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &order, nil
}

func (r *orderRepository) Update(ctx context.Context, order *domain.Order) error {
	// 与 GORM 的 Save 相同，但逻辑上更明确
	return r.Save(ctx, order)
}

func (r *orderRepository) Delete(ctx context.Context, id uint) error {
	db := r.sharding.GetDB(0) // TODO: Sharding support
	return db.WithContext(ctx).Delete(&domain.Order{}, id).Error
}

func (r *orderRepository) List(ctx context.Context, offset, limit int) ([]*domain.Order, int64, error) {
	// Scan all shards? Or just shard 0?
	// 目前，在分片中列出没有 UserID 的所有订单开销很大。
	// 我们默认为分片 0，或者可能需要遍历所有分片（此处为简单起见未实现）。
	db := r.sharding.GetDB(0).WithContext(ctx).Model(&domain.Order{})

	var list []*domain.Order
	var total int64

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := db.Preload("Items").Offset(offset).Limit(limit).Order("created_at desc").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

func (r *orderRepository) ListByUserID(ctx context.Context, userID uint, offset, limit int) ([]*domain.Order, int64, error) {
	db := r.sharding.GetDB(uint64(userID)).WithContext(ctx).Model(&domain.Order{})

	var list []*domain.Order
	var total int64

	db = db.Where("user_id = ?", userID)

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := db.Preload("Items").Offset(offset).Limit(limit).Order("created_at desc").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}
