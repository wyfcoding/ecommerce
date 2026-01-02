package persistence

import (
	"context"
	"errors"
	"sort"

	"github.com/wyfcoding/ecommerce/internal/order/domain"
	"github.com/wyfcoding/pkg/databases/sharding"

	"gorm.io/gorm"
)

type orderRepository struct {
	sharding *sharding.Manager
	tx       *gorm.DB
}

// NewOrderRepository 定义了数据持久层接口。
func NewOrderRepository(sharding *sharding.Manager) domain.OrderRepository {
	return &orderRepository{sharding: sharding}
}

// WithTx 实现了 domain.OrderRepository 接口，支持事务嵌套。
func (r *orderRepository) WithTx(tx any) domain.OrderRepository {
	if gormTx, ok := tx.(*gorm.DB); ok {
		return &orderRepository{
			sharding: r.sharding,
			tx:       gormTx,
		}
	}
	return r
}

// getDB 内部辅助方法，自动切换事务与普通连接
func (r *orderRepository) getDB(userID uint64) *gorm.DB {
	if r.tx != nil {
		return r.tx
	}
	return r.sharding.GetDB(userID)
}

// Save 将订单聚合根保存到对应的分库中。
func (r *orderRepository) Save(ctx context.Context, order *domain.Order) error {
	db := r.getDB(order.UserID)

	execute := func(tx *gorm.DB) error {
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
	}

	if r.tx != nil {
		return execute(r.tx.WithContext(ctx))
	}

	return db.WithContext(ctx).Transaction(execute)
}

// Transaction 实现了事务包装逻辑
func (r *orderRepository) Transaction(ctx context.Context, userID uint64, fn func(tx any) error) error {
	db := r.sharding.GetDB(userID)
	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(tx)
	})
}

// FindByID 根据ID从数据库中查询订单。
func (r *orderRepository) FindByID(ctx context.Context, userID uint64, id uint) (*domain.Order, error) {
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

// Delete 根据ID物理删除订单记录。
func (r *orderRepository) Delete(ctx context.Context, userID uint64, id uint) error {
	db := r.sharding.GetDB(userID)
	return db.WithContext(ctx).Delete(&domain.Order{}, id).Error
}

// List 全局分页列出所有订单记录。
func (r *orderRepository) List(ctx context.Context, offset, limit int) ([]*domain.Order, int64, error) {
	dbs := r.sharding.GetAllDBs()
	var allOrders []*domain.Order
	var totalCount int64

	// 真实化执行：分布式全表聚合与全局排序分页
	for _, db := range dbs {
		var list []*domain.Order
		var count int64
		query := db.WithContext(ctx).Model(&domain.Order{})
		if err := query.Count(&count).Error; err != nil {
			return nil, 0, err
		}
		totalCount += count

		// 为保证分页准确性，每个分片拉取足够的样本进行合并排序
		if err := query.Preload("Items").Order("created_at desc").Limit(offset + limit).Find(&list).Error; err != nil {
			return nil, 0, err
		}
		allOrders = append(allOrders, list...)
	}

	// 全局按时间降序排列
	sort.Slice(allOrders, func(i, j int) bool {
		return allOrders[i].CreatedAt.After(allOrders[j].CreatedAt)
	})

	start := offset
	if start > len(allOrders) {
		return []*domain.Order{}, totalCount, nil
	}
	end := offset + limit
	if end > len(allOrders) {
		end = len(allOrders)
	}

	return allOrders[start:end], totalCount, nil
}

// ListByUserID 获取指定用户的订单列表。
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
