// 变更说明：迁移至 mysql 子目录，明确存储技术边界。
// 假设：该仓储仅负责 MySQL 写模型，不直接承担读模型查询。
package mysql

import (
	"context"
	"errors"
	"slices"

	"github.com/wyfcoding/ecommerce/internal/order/domain"
	"github.com/wyfcoding/pkg/database/sharding"

	"gorm.io/gorm"
)

type orderRepository struct {
	sharding *sharding.Manager
}

// NewOrderRepository 定义了数据持久层接口。
func NewOrderRepository(sharding *sharding.Manager) domain.OrderRepository {
	return &orderRepository{sharding: sharding}
}

// BeginTx 开始事务 (基于 UserID 定位分库)
func (r *orderRepository) BeginTx(ctx context.Context, userID uint64) any {
	return r.sharding.GetDB(userID).WithContext(ctx).Begin()
}

// CommitTx 提交事务
func (r *orderRepository) CommitTx(tx any) error {
	return tx.(*gorm.DB).Commit().Error
}

// RollbackTx 回滚事务
func (r *orderRepository) RollbackTx(tx any) error {
	return tx.(*gorm.DB).Rollback().Error
}

// WithTx 事务包装器
func (r *orderRepository) WithTx(ctx context.Context, userID uint64, fn func(tx any) error) error {
	db := r.sharding.GetDB(userID)
	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(tx)
	})
}

// Save 将订单聚合根保存到对应的分库中。
func (r *orderRepository) Save(ctx context.Context, order *domain.Order) error {
	db := r.sharding.GetDB(order.UserID).WithContext(ctx)
	return db.Transaction(func(tx *gorm.DB) error {
		return r.SaveInTx(ctx, tx, order)
	})
}

// SaveInTx 在事务中保存订单聚合根。
func (r *orderRepository) SaveInTx(ctx context.Context, tx any, order *domain.Order) error {
	gormTx := tx.(*gorm.DB).WithContext(ctx)
	if err := gormTx.Save(order).Error; err != nil {
		return err
	}
	for _, item := range order.Items {
		if item.ID == 0 {
			item.OrderID = uint64(order.ID)
		}
		if err := gormTx.Save(item).Error; err != nil {
			return err
		}
	}
	for _, log := range order.Logs {
		if log.ID == 0 {
			log.OrderID = uint64(order.ID)
		}
		if err := gormTx.Save(log).Error; err != nil {
			return err
		}
	}
	return nil
}

// FindByID 根据ID从数据库中查询订单。
func (r *orderRepository) FindByID(ctx context.Context, userID uint64, id uint64) (*domain.Order, error) {
	db := r.sharding.GetDB(userID)
	var order domain.Order
	if err := db.WithContext(ctx).Preload("Items").Preload("Logs").First(&order, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &order, nil
}

// FindByOrderNo 根据订单编号查询订单。
func (r *orderRepository) FindByOrderNo(ctx context.Context, userID uint64, orderNo string) (*domain.Order, error) {
	db := r.sharding.GetDB(userID)
	var order domain.Order
	if err := db.WithContext(ctx).Preload("Items").Preload("Logs").Where("order_no = ?", orderNo).First(&order).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &order, nil
}

// Update 更新订单聚合根状态及相关信息。
func (r *orderRepository) Update(ctx context.Context, order *domain.Order) error {
	return r.Save(ctx, order)
}

// UpdateInTx 在事务中更新。
func (r *orderRepository) UpdateInTx(ctx context.Context, tx any, order *domain.Order) error {
	return r.SaveInTx(ctx, tx, order)
}

// Delete 根据ID物理删除订单记录。
func (r *orderRepository) Delete(ctx context.Context, userID uint64, id uint64) error {
	db := r.sharding.GetDB(userID)
	return db.WithContext(ctx).Delete(&domain.Order{}, id).Error
}

// List 全局分页列出所有订单记录。
func (r *orderRepository) List(ctx context.Context, offset, limit int) ([]*domain.Order, int64, error) {
	dbs := r.sharding.GetAllDBs()
	var allOrders []*domain.Order
	var totalCount int64

	for _, db := range dbs {
		var list []*domain.Order
		var count int64
		query := db.WithContext(ctx).Model(&domain.Order{})
		if err := query.Count(&count).Error; err != nil {
			return nil, 0, err
		}
		totalCount += count

		// 获取样本
		if err := query.Preload("Items").Order("created_at desc").Limit(offset + limit).Find(&list).Error; err != nil {
			return nil, 0, err
		}
		allOrders = append(allOrders, list...)
	}

	// 全局排序
	slices.SortFunc(allOrders, func(a, b *domain.Order) int {
		if a.CreatedAt.After(b.CreatedAt) {
			return -1
		}
		if b.CreatedAt.After(a.CreatedAt) {
			return 1
		}
		return 0
	})

	start := offset
	if start > len(allOrders) {
		return []*domain.Order{}, totalCount, nil
	}
	end := min(offset+limit, len(allOrders))

	return allOrders[start:end], totalCount, nil
}

// ListByUserID 获取指定用户的订单列表。
func (r *orderRepository) ListByUserID(ctx context.Context, userID uint64, status *int, offset, limit int) ([]*domain.Order, int64, error) {
	db := r.sharding.GetDB(userID).WithContext(ctx).Model(&domain.Order{})

	db = db.Where("user_id = ?", userID)
	if status != nil {
		db = db.Where("status = ?", *status)
	}

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
