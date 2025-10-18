package repository

import (
	"context"
	"fmt"
	"time"

	"ecommerce/internal/order/model"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// OrderRepo 定义了订单数据的存储接口。
type OrderRepo interface {
	// CreateOrder 创建一个新的订单，包括订单项和收货地址。
	CreateOrder(ctx context.Context, order *model.Order) (*model.Order, error)
	// GetOrderByID 根据ID获取订单详情，包含订单项、收货地址和日志。
	GetOrderByID(ctx context.Context, id uint64) (*model.Order, error)
	// UpdateOrder 更新订单信息。
	UpdateOrder(ctx context.Context, order *model.Order) (*model.Order, error)
	// DeleteOrder 逻辑删除订单。
	DeleteOrder(ctx context.Context, id uint64) error
	// ListOrders 根据条件分页查询订单列表。
	ListOrders(ctx context.Context, query *OrderListQuery) ([]*model.Order, int64, error)
	// GetOrderByOrderNo 根据订单编号获取订单。
	GetOrderByOrderNo(ctx context.Context, orderNo string) (*model.Order, error)
}

// OrderItemRepo 定义了订单项数据的存储接口。
type OrderItemRepo interface {
	// CreateOrderItems 批量创建订单项。
	CreateOrderItems(ctx context.Context, items []*model.OrderItem) ([]*model.OrderItem, error)
	// GetOrderItemsByOrderID 根据订单ID获取所有订单项。
	GetOrderItemsByOrderID(ctx context.Context, orderID uint64) ([]*model.OrderItem, error)
}

// ShippingAddressRepo 定义了收货地址数据的存储接口。
type ShippingAddressRepo interface {
	// CreateShippingAddress 创建收货地址。
	CreateShippingAddress(ctx context.Context, address *model.ShippingAddress) (*model.ShippingAddress, error)
	// GetShippingAddressByOrderID 根据订单ID获取收货地址。
	GetShippingAddressByOrderID(ctx context.Context, orderID uint64) (*model.ShippingAddress, error)
	// UpdateShippingAddress 更新收货地址。
	UpdateShippingAddress(ctx context.Context, address *model.ShippingAddress) (*model.ShippingAddress, error)
}

// OrderLogRepo 定义了订单日志数据的存储接口。
type OrderLogRepo interface {
	// CreateOrderLog 创建订单日志。
	CreateOrderLog(ctx context.Context, log *model.OrderLog) (*model.OrderLog, error)
	// ListOrderLogsByOrderID 根据订单ID获取所有订单日志。
	ListOrderLogsByOrderID(ctx context.Context, orderID uint64) ([]*model.OrderLog, error)
}

// OrderListQuery 定义订单列表查询的参数。
type OrderListQuery struct {
	Page      int32
	PageSize  int32
	UserID    uint64
	Status    model.OrderStatus
	StartTime *time.Time
	EndTime   *time.Time
	SortBy    string // 例如: "created_at_desc", "total_amount_asc"
}

// orderRepoImpl 是 OrderRepo 接口的 GORM 实现。
type orderRepoImpl struct {
	db *gorm.DB
}

// NewOrderRepo 创建一个新的 OrderRepo 实例。
func NewOrderRepo(db *gorm.DB) OrderRepo {
	return &orderRepoImpl{db: db}
}

// CreateOrder 实现 OrderRepo.CreateOrder 方法。
func (r *orderRepoImpl) CreateOrder(ctx context.Context, order *model.Order) (*model.Order, error) {
	// 使用事务确保订单、订单项和收货地址的原子性创建
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(order).Error; err != nil {
			return fmt.Errorf("failed to create order: %w", err)
		}
		// 关联订单项
		for i := range order.Items {
			order.Items[i].OrderID = order.ID
		}
		if len(order.Items) > 0 {
			if err := tx.Create(&order.Items).Error; err != nil {
				return fmt.Errorf("failed to create order items: %w", err)
			}
		}
		// 关联收货地址
		if order.ShippingAddress.RecipientName != "" { // 简单判断地址是否为空
			order.ShippingAddress.OrderID = order.ID
			if err := tx.Create(&order.ShippingAddress).Error; err != nil {
				return fmt.Errorf("failed to create shipping address: %w", err)
			}
		}
		// 关联订单日志
		for i := range order.Logs {
			order.Logs[i].OrderID = order.ID
		}
		if len(order.Logs) > 0 {
			if err := tx.Create(&order.Logs).Error; err != nil {
				return fmt.Errorf("failed to create order logs: %w", err)
			}
		}
		return nil
	})

	if err != nil {
		zap.S().Errorf("failed to create order: %v", err)
		return nil, err
	}
	return order, nil
}

// GetOrderByID 实现 OrderRepo.GetOrderByID 方法。
func (r *orderRepoImpl) GetOrderByID(ctx context.Context, id uint64) (*model.Order, error) {
	var order model.Order
	// 预加载所有关联数据
	if err := r.db.WithContext(ctx).Preload("Items").Preload("ShippingAddress").Preload("Logs").First(&order, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		zap.S().Errorf("failed to get order by id %d: %v", id, err)
		return nil, fmt.Errorf("failed to get order by id: %w", err)
	}
	return &order, nil
}

// UpdateOrder 实现 OrderRepo.UpdateOrder 方法。
func (r *orderRepoImpl) UpdateOrder(ctx context.Context, order *model.Order) (*model.Order, error) {
	// 仅更新订单主表信息，关联数据需要单独更新或通过事务管理
	if err := r.db.WithContext(ctx).Save(order).Error; err != nil {
		zap.S().Errorf("failed to update order %d: %v", order.ID, err)
		return nil, fmt.Errorf("failed to update order: %w", err)
	}
	return order, nil
}

// DeleteOrder 实现 OrderRepo.DeleteOrder 方法 (逻辑删除)。
func (r *orderRepoImpl) DeleteOrder(ctx context.Context, id uint64) error {
	// GORM 的 Delete 方法默认执行软删除 (如果模型包含 gorm.DeletedAt 字段)
	if err := r.db.WithContext(ctx).Delete(&model.Order{}, id).Error; err != nil {
		zap.S().Errorf("failed to delete order %d: %v", id, err)
		return fmt.Errorf("failed to delete order: %w", err)
	}
	return nil
}

// ListOrders 实现 OrderRepo.ListOrders 方法。
func (r *orderRepoImpl) ListOrders(ctx context.Context, query *OrderListQuery) ([]*model.Order, int64, error) {
	var orders []*model.Order
	var total int64

	db := r.db.WithContext(ctx).Model(&model.Order{}).Preload("Items").Preload("ShippingAddress")

	// 应用筛选条件
	if query.UserID != 0 {
		db = db.Where("user_id = ?", query.UserID)
	}
	// 状态筛选，排除未指定状态
	if query.Status != model.OrderStatusUnspecified {
		db = db.Where("status = ?", query.Status)
	}
	if query.StartTime != nil {
		db = db.Where("created_at >= ?", *query.StartTime)
	}
	if query.EndTime != nil {
		db = db.Where("created_at <= ?", *query.EndTime)
	}

	// 统计总数
	if err := db.Count(&total).Error; err != nil {
		zap.S().Errorf("failed to count orders: %v", err)
		return nil, 0, fmt.Errorf("failed to count orders: %w", err)
	}

	// 应用排序
	if query.SortBy != "" {
		db = db.Order(query.SortBy)
	} else {
		db = db.Order("created_at DESC") // 默认按创建时间降序
	}

	// 应用分页
	if query.PageSize > 0 && query.Page > 0 {
		offset := (query.Page - 1) * query.PageSize
		db = db.Limit(int(query.PageSize)).Offset(int(offset))
	}

	// 查询数据
	if err := db.Find(&orders).Error; err != nil {
		zap.S().Errorf("failed to list orders: %v", err)
		return nil, 0, fmt.Errorf("failed to list orders: %w", err)
	}

	return orders, total, nil
}

// GetOrderByOrderNo 实现 OrderRepo.GetOrderByOrderNo 方法。
func (r *orderRepoImpl) GetOrderByOrderNo(ctx context.Context, orderNo string) (*model.Order, error) {
	var order model.Order
	if err := r.db.WithContext(ctx).Where("order_no = ?", orderNo).Preload("Items").Preload("ShippingAddress").Preload("Logs").First(&order).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		zap.S().Errorf("failed to get order by order_no %s: %v", orderNo, err)
		return nil, fmt.Errorf("failed to get order by order_no: %w", err)
	}
	return &order, nil
}

// orderItemRepoImpl 是 OrderItemRepo 接口的 GORM 实现。
type orderItemRepoImpl struct {
	db *gorm.DB
}

// NewOrderItemRepo 创建一个新的 OrderItemRepo 实例。
func NewOrderItemRepo(db *gorm.DB) OrderItemRepo {
	return &orderItemRepoImpl{db: db}
}

// CreateOrderItems 实现 OrderItemRepo.CreateOrderItems 方法。
func (r *orderItemRepoImpl) CreateOrderItems(ctx context.Context, items []*model.OrderItem) ([]*model.OrderItem, error) {
	if err := r.db.WithContext(ctx).Create(&items).Error; err != nil {
		zap.S().Errorf("failed to create order items: %v", err)
		return nil, fmt.Errorf("failed to create order items: %w", err)
	}
	return items, nil
}

// GetOrderItemsByOrderID 实现 OrderItemRepo.GetOrderItemsByOrderID 方法。
func (r *orderItemRepoImpl) GetOrderItemsByOrderID(ctx context.Context, orderID uint64) ([]*model.OrderItem, error) {
	var items []*model.OrderItem
	if err := r.db.WithContext(ctx).Where("order_id = ?", orderID).Find(&items).Error; err != nil {
		zap.S().Errorf("failed to get order items for order %d: %v", orderID, err)
		return nil, fmt.Errorf("failed to get order items by order id: %w", err)
	}
	return items, nil
}

// shippingAddressRepoImpl 是 ShippingAddressRepo 接口的 GORM 实现。
type shippingAddressRepoImpl struct {
	db *gorm.DB
}

// NewShippingAddressRepo 创建一个新的 ShippingAddressRepo 实例。
func NewShippingAddressRepo(db *gorm.DB) ShippingAddressRepo {
	return &shippingAddressRepoImpl{db: db}
}

// CreateShippingAddress 实现 ShippingAddressRepo.CreateShippingAddress 方法。
func (r *shippingAddressRepoImpl) CreateShippingAddress(ctx context.Context, address *model.ShippingAddress) (*model.ShippingAddress, error) {
	if err := r.db.WithContext(ctx).Create(address).Error; err != nil {
		zap.S().Errorf("failed to create shipping address: %v", err)
		return nil, fmt.Errorf("failed to create shipping address: %w", err)
	}
	return address, nil
}

// GetShippingAddressByOrderID 实现 ShippingAddressRepo.GetShippingAddressByOrderID 方法。
func (r *shippingAddressRepoImpl) GetShippingAddressByOrderID(ctx context.Context, orderID uint64) (*model.ShippingAddress, error) {
	var address model.ShippingAddress
	if err := r.db.WithContext(ctx).Where("order_id = ?", orderID).First(&address).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		zap.S().Errorf("failed to get shipping address for order %d: %v", orderID, err)
		return nil, fmt.Errorf("failed to get shipping address by order id: %w", err)
	}
	return &address, nil
}

// UpdateShippingAddress 实现 ShippingAddressRepo.UpdateShippingAddress 方法。
func (r *shippingAddressRepoImpl) UpdateShippingAddress(ctx context.Context, address *model.ShippingAddress) (*model.ShippingAddress, error) {
	if err := r.db.WithContext(ctx).Save(address).Error; err != nil {
		zap.S().Errorf("failed to update shipping address %d: %v", address.ID, err)
		return nil, fmt.Errorf("failed to update shipping address: %w", err)
	}
	return address, nil
}

// orderLogRepoImpl 是 OrderLogRepo 接口的 GORM 实现。
type orderLogRepoImpl struct {
	db *gorm.DB
}

// NewOrderLogRepo 创建一个新的 OrderLogRepo 实例。
func NewOrderLogRepo(db *gorm.DB) OrderLogRepo {
	return &orderLogRepoImpl{db: db}
}

// CreateOrderLog 实现 OrderLogRepo.CreateOrderLog 方法。
func (r *orderLogRepoImpl) CreateOrderLog(ctx context.Context, log *model.OrderLog) (*model.OrderLog, error) {
	if err := r.db.WithContext(ctx).Create(log).Error; err != nil {
		zap.S().Errorf("failed to create order log: %v", err)
		return nil, fmt.Errorf("failed to create order log: %w", err)
	}
	return log, nil
}

// ListOrderLogsByOrderID 实现 OrderLogRepo.ListOrderLogsByOrderID 方法。
func (r *orderLogRepoImpl) ListOrderLogsByOrderID(ctx context.Context, orderID uint64) ([]*model.OrderLog, error) {
	var logs []*model.OrderLog
	if err := r.db.WithContext(ctx).Where("order_id = ?", orderID).Order("created_at ASC").Find(&logs).Error; err != nil {
		zap.S().Errorf("failed to list order logs for order %d: %v", orderID, err)
		return nil, fmt.Errorf("failed to list order logs by order id: %w", err)
	}
	return logs, nil
}
