// 变更说明：迁移至 mysql 子目录，明确存储技术边界。
// 假设：该仓储仅负责 MySQL 写模型，不直接承担读模型查询。
package mysql

import (
	"context"
	"errors"
	"slices"
	"strings"
	"time"

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
	model := toOrderModel(order)
	if err := gormTx.Save(model).Error; err != nil {
		return err
	}
	for i := range model.Items {
		if model.Items[i].ID == 0 {
			model.Items[i].OrderID = uint64(model.ID)
		}
		if err := gormTx.Save(&model.Items[i]).Error; err != nil {
			return err
		}
	}
	for i := range model.Logs {
		if model.Logs[i].ID == 0 {
			model.Logs[i].OrderID = uint64(model.ID)
		}
		if err := gormTx.Save(&model.Logs[i]).Error; err != nil {
			return err
		}
	}

	if order != nil {
		if synced := toDomainOrder(model); synced != nil {
			*order = *synced
		}
	}
	return nil
}

// FindByID 根据ID从数据库中查询订单。
func (r *orderRepository) FindByID(ctx context.Context, userID uint64, id uint64) (*domain.Order, error) {
	db := r.sharding.GetDB(userID)
	var order OrderModel
	if err := db.WithContext(ctx).Preload("Items").Preload("Logs").First(&order, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toDomainOrder(&order), nil
}

// FindByOrderNo 根据订单编号查询订单。
func (r *orderRepository) FindByOrderNo(ctx context.Context, userID uint64, orderNo string) (*domain.Order, error) {
	db := r.sharding.GetDB(userID)
	var order OrderModel
	if err := db.WithContext(ctx).Preload("Items").Preload("Logs").Where("order_no = ?", orderNo).First(&order).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toDomainOrder(&order), nil
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
	return db.WithContext(ctx).Delete(&OrderModel{}, id).Error
}

// List 全局分页列出所有订单记录。
func (r *orderRepository) List(ctx context.Context, status *int, offset, limit int, startTime, endTime *time.Time, sortBy string) ([]*domain.Order, int64, error) {
	dbs := r.sharding.GetAllDBs()
	var allOrders []*domain.Order
	var totalCount int64
	orderColumn, desc := parseOrderSort(sortBy)

	for _, db := range dbs {
		var list []OrderModel
		var count int64
		query := db.WithContext(ctx).Model(&OrderModel{})
		query = applyOrderFilters(query, status, startTime, endTime)
		if err := query.Count(&count).Error; err != nil {
			return nil, 0, err
		}
		totalCount += count

		// 获取样本
		orderClause := orderColumn
		if desc {
			orderClause += " desc"
		} else {
			orderClause += " asc"
		}
		if err := query.Preload("Items").Preload("Logs").Order(orderClause).Limit(offset + limit).Find(&list).Error; err != nil {
			return nil, 0, err
		}
		for i := range list {
			order := toDomainOrder(&list[i])
			if order != nil {
				allOrders = append(allOrders, order)
			}
		}
	}

	// 全局排序
	slices.SortFunc(allOrders, func(a, b *domain.Order) int {
		cmp := compareOrderByColumn(a, b, orderColumn)
		if desc {
			cmp = -cmp
		}
		return cmp
	})

	start := offset
	if start > len(allOrders) {
		return []*domain.Order{}, totalCount, nil
	}
	end := min(offset+limit, len(allOrders))

	return allOrders[start:end], totalCount, nil
}

// ListByUserID 获取指定用户的订单列表。
func (r *orderRepository) ListByUserID(ctx context.Context, userID uint64, status *int, offset, limit int, startTime, endTime *time.Time, sortBy string) ([]*domain.Order, int64, error) {
	db := r.sharding.GetDB(userID).WithContext(ctx).Model(&OrderModel{})

	db = db.Where("user_id = ?", userID)
	db = applyOrderFilters(db, status, startTime, endTime)

	var list []OrderModel
	var total int64

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	orderColumn, desc := parseOrderSort(sortBy)
	orderClause := orderColumn
	if desc {
		orderClause += " desc"
	} else {
		orderClause += " asc"
	}
	if err := db.Preload("Items").Preload("Logs").Offset(offset).Limit(limit).Order(orderClause).Find(&list).Error; err != nil {
		return nil, 0, err
	}

	orders := make([]*domain.Order, 0, len(list))
	for i := range list {
		order := toDomainOrder(&list[i])
		if order != nil {
			orders = append(orders, order)
		}
	}
	return orders, total, nil
}

func applyOrderFilters(db *gorm.DB, status *int, startTime, endTime *time.Time) *gorm.DB {
	if status != nil {
		db = db.Where("status = ?", *status)
	}
	if startTime != nil {
		db = db.Where("created_at >= ?", *startTime)
	}
	if endTime != nil {
		db = db.Where("created_at <= ?", *endTime)
	}
	return db
}

func parseOrderSort(sortBy string) (string, bool) {
	allowed := map[string]string{
		"created_at":    "created_at",
		"updated_at":    "updated_at",
		"paid_at":       "paid_at",
		"shipped_at":    "shipped_at",
		"delivered_at":  "delivered_at",
		"completed_at":  "completed_at",
		"cancelled_at":  "cancelled_at",
		"total_amount":  "total_amount",
		"actual_amount": "actual_amount",
	}

	sortBy = strings.TrimSpace(strings.ToLower(sortBy))
	if sortBy == "" {
		return "created_at", true
	}

	desc := true
	if strings.HasPrefix(sortBy, "-") {
		sortBy = strings.TrimPrefix(sortBy, "-")
		desc = true
	}

	parts := strings.Fields(sortBy)
	if len(parts) > 0 {
		sortBy = parts[0]
	}
	if len(parts) > 1 {
		switch parts[1] {
		case "asc":
			desc = false
		case "desc":
			desc = true
		}
	}

	if col, ok := allowed[sortBy]; ok {
		return col, desc
	}
	return "created_at", true
}

func compareOrderByColumn(a, b *domain.Order, column string) int {
	switch column {
	case "updated_at":
		return compareTime(a.UpdatedAt, b.UpdatedAt)
	case "paid_at":
		return compareTime(timeOrZero(a.PaidAt), timeOrZero(b.PaidAt))
	case "shipped_at":
		return compareTime(timeOrZero(a.ShippedAt), timeOrZero(b.ShippedAt))
	case "delivered_at":
		return compareTime(timeOrZero(a.DeliveredAt), timeOrZero(b.DeliveredAt))
	case "completed_at":
		return compareTime(timeOrZero(a.CompletedAt), timeOrZero(b.CompletedAt))
	case "cancelled_at":
		return compareTime(timeOrZero(a.CancelledAt), timeOrZero(b.CancelledAt))
	case "total_amount":
		return compareInt64(a.TotalAmount, b.TotalAmount)
	case "actual_amount":
		return compareInt64(a.ActualAmount, b.ActualAmount)
	default:
		return compareTime(a.CreatedAt, b.CreatedAt)
	}
}

func timeOrZero(t *time.Time) time.Time {
	if t == nil {
		return time.Time{}
	}
	return *t
}

func compareTime(a, b time.Time) int {
	if a.Before(b) {
		return -1
	}
	if a.After(b) {
		return 1
	}
	return 0
}

func compareInt64(a, b int64) int {
	if a < b {
		return -1
	}
	if a > b {
		return 1
	}
	return 0
}
