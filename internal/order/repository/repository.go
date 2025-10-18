package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"ecommerce/internal/order/model"
)

// OrderRepository 定义了订单数据仓库的接口
type OrderRepository interface {
	CreateOrder(ctx context.Context, order *model.Order) error
	GetOrderByID(ctx context.Context, id uint) (*model.Order, error)
	GetOrderByOrderSN(ctx context.Context, orderSN string) (*model.Order, error)
	ListOrdersByUserID(ctx context.Context, userID uint, page, pageSize int) ([]model.Order, int64, error)
	UpdateOrderStatus(ctx context.Context, orderID uint, status model.OrderStatus) error
	UpdateOrder(ctx context.Context, order *model.Order) error
}

// orderRepository 是接口的具体实现
type orderRepository struct {
	db *gorm.DB
}

// NewOrderRepository 创建一个新的 orderRepository 实例
func NewOrderRepository(db *gorm.DB) OrderRepository {
	return &orderRepository{db: db}
}

// CreateOrder 在数据库中创建订单和订单项
// 这个操作必须是事务性的，以保证数据一致性
func (r *orderRepository) CreateOrder(ctx context.Context, order *model.Order) error {
	// 开始一个数据库事务
	tx := r.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return fmt.Errorf("开启数据库事务失败: %w", tx.Error)
	}

	// 在事务中执行操作
	// 1. 创建订单主记录
	if err := tx.Create(order).Error; err != nil {
		tx.Rollback() // 如果出错，回滚事务
		return fmt.Errorf("数据库创建订单失败: %w", err)
	}

	// 2. 循环创建订单项记录 (order.Items 应该在 service 层被填充)
	// GORM 的关联创建会自动处理这个，但为了清晰，我们也可以手动处理
	// if err := tx.Create(&order.Items).Error; err != nil {
	// 	tx.Rollback()
	// 	return fmt.Errorf("数据库创建订单项失败: %w", err)
	// }

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("提交数据库事务失败: %w", err)
	}

	return nil
}

// GetOrderByID 根据主键 ID 获取订单
func (r *orderRepository) GetOrderByID(ctx context.Context, id uint) (*model.Order, error) {
	var order model.Order
	// Preload("Items") 会同时加载订单中的所有商品项
	if err := r.db.WithContext(ctx).Preload("Items").First(&order, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // 订单不存在
		}
		return nil, fmt.Errorf("数据库查询订单失败: %w", err)
	}
	return &order, nil
}

// GetOrderByOrderSN 根据业务订单号获取订单
func (r *orderRepository) GetOrderByOrderSN(ctx context.Context, orderSN string) (*model.Order, error) {
	var order model.Order
	if err := r.db.WithContext(ctx).Preload("Items").Where("order_sn = ?", orderSN).First(&order).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("数据库查询订单失败: %w", err)
	}
	return &order, nil
}

// ListOrdersByUserID 分页列出某个用户的所有订单
func (r *orderRepository) ListOrdersByUserID(ctx context.Context, userID uint, page, pageSize int) ([]model.Order, int64, error) {
	var orders []model.Order
	var total int64

	db := r.db.WithContext(ctx).Model(&model.Order{}).Where("user_id = ?", userID)

	// 计算总数
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("数据库统计订单数量失败: %w", err)
	}

	// 计算偏移量并查询
	offset := (page - 1) * pageSize
	// 按创建时间降序排序，让用户先看到最新的订单
	if err := db.Order("created_at desc").Offset(offset).Limit(pageSize).Find(&orders).Error; err != nil {
		return nil, 0, fmt.Errorf("数据库列出订单失败: %w", err)
	}

	return orders, total, nil
}

// UpdateOrderStatus 更新订单的状态
func (r *orderRepository) UpdateOrderStatus(ctx context.Context, orderID uint, status model.OrderStatus) error {
	result := r.db.WithContext(ctx).Model(&model.Order{}).Where("id = ?", orderID).Update("status", status)
	if result.Error != nil {
		return fmt.Errorf("数据库更新订单状态失败: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("订单不存在或状态未改变")
	}
	return nil
}

// UpdateOrder 更新整个订单信息 (例如，支付成功后更新支付信息)
func (r *orderRepository) UpdateOrder(ctx context.Context, order *model.Order) error {
	if err := r.db.WithContext(ctx).Save(order).Error; err != nil {
		return fmt.Errorf("数据库更新订单信息失败: %w", err)
	}
	return nil
}
