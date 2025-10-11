
package model

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	cartv1 "ecommerce/api/cart/v1"
	paymentv1 "ecommerce/api/payment/v1"
	productv1 "ecommerce/api/product/v1"
	"ecommerce/pkg/snowflake"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// 定义订单相关的错误
var (
	ErrOrderNotFound     = errors.New("order not found")
	ErrInvalidStatus     = errors.New("invalid order status")
	ErrStockNotEnough    = errors.New("stock not enough")
	ErrOrderNotPending   = errors.New("order is not in pending status")
	ErrOrderAlreadyPaid  = errors.New("order has already been paid")
	ErrOrderBelongsToAnotherUser = errors.New("order does not belong to user")
)

// OrderStatus 定义订单状态
type OrderStatus int32

const (
	OrderStatusUnknown    OrderStatus = 0 // 未知状态
	OrderStatusPending    OrderStatus = 1 // 待支付
	OrderStatusPaid       OrderStatus = 2 // 已支付
	OrderStatusShipped    OrderStatus = 3 // 已发货
	OrderStatusCompleted  OrderStatus = 4 // 已完成
	OrderStatusCancelled  OrderStatus = 5 // 已取消
	OrderStatusRefunded   OrderStatus = 6 // 已退款
)

// String 方法用于方便地打印订单状态
func (s OrderStatus) String() string {
	switch s {
	case OrderStatusPending:
		return "Pending"
	case OrderStatusPaid:
		return "Paid"
	case OrderStatusShipped:
		return "Shipped"
	case OrderStatusCompleted:
		return "Completed"
	case OrderStatusCancelled:
		return "Cancelled"
	case OrderStatusRefunded:
		return "Refunded"
	default:
		return "Unknown"
	}
}

// Order 是订单的业务领域模型和数据库模型
type Order struct {
	gorm.Model
	ID            uint64      `gorm:"primaryKey"`
	UserID        uint64      `gorm:"not null;index"`
	TotalAmount   float64     `gorm:"type:decimal(10,2);not null"`
	PaymentAmount float64     `gorm:"type:decimal(10,2);not null"`
	ShippingFee   float64     `gorm:"type:decimal(10,2);not null"`
	Status        OrderStatus `gorm:"not null;default:0"` // 订单状态
	AddressID     uint64      `gorm:"not null"`
	Remark        string      `gorm:"type:varchar(255)"`
	OrderItems    []OrderItem `gorm:"foreignKey:OrderID"` // 订单项
}

// OrderItem 是订单项的业务领域模型和数据库模型
type OrderItem struct {
	gorm.Model
	OrderID      uint64  `gorm:"not null;index"`
	SKUID        uint64  `gorm:"not null"`
	ProductName  string  `gorm:"type:varchar(255);not null"`
	ProductImage string  `gorm:"type:varchar(255)"`
	Price        float64 `gorm:"type:decimal(10,2);not null"`
	Quantity     uint32  `gorm:"not null"`
}

// PaymentInfo 包含支付相关信息
type PaymentInfo struct {
	PaymentID  string
	PaymentURL string
}

// OrderRepo 定义了订单数据仓库需要实现的接口
type OrderRepo interface {
	CreateOrder(ctx context.Context, order *Order, items []*OrderItem) (*Order, error)
	GetOrderByID(ctx context.Context, orderID uint64) (*Order, error)
	UpdateOrderStatus(ctx context.Context, orderID uint64, status OrderStatus) error
	// ... 其他订单相关的持久化操作
}

// ProductClient 定义了与商品服务交互的接口
type ProductClient interface {
	GetProductSKU(ctx context.Context, skuID uint64) (*productv1.SkuInfo, error)
	LockStock(ctx context.Context, skuID uint64, quantity uint32) error
	UnlockStock(ctx context.Context, skuID uint64, quantity uint32) error
}

// CartClient 定义了与购物车服务交互的接口
type CartClient interface {
	ClearCart(ctx context.Context, userID uint64) error
}

// PaymentClient 定义了与支付服务交互的接口
type PaymentClient interface {
	CreatePayment(ctx context.Context, userID uint64, orderID string, amount float64) (*PaymentInfo, error)
	ProcessPaymentNotification(ctx context.Context, paymentID string, data map[string]string) error
}

// Transaction 定义了事务管理器接口
type Transaction interface {
	InTx(ctx context.Context, fn func(ctx context.Context) error) error
}

// OrderUsecase 是订单的业务用例
type OrderUsecase struct {
	repo          OrderRepo
	productClient ProductClient
	cartClient    CartClient
	paymentClient PaymentClient
	transaction   Transaction
	logger        *zap.SugaredLogger
}

// NewOrderUsecase 是 OrderUsecase 的构造函数
func NewOrderUsecase(repo OrderRepo, productClient ProductClient, cartClient CartClient, paymentClient PaymentClient, transaction Transaction, logger *zap.SugaredLogger) *OrderUsecase {
	return &OrderUsecase{
		repo:          repo,
		productClient: productClient,
		cartClient:    cartClient,
		paymentClient: paymentClient,
		transaction:   transaction,
		logger:        logger,
	}
}

// CreateOrderRequestItem 定义了创建订单时每个商品项的输入结构
type CreateOrderRequestItem struct {
	SkuID    uint64
	Quantity uint32
}

// CreateOrder 是创建订单的核心业务流程
// 它通过事务确保了订单创建、库存锁定等操作的原子性
func (uc *OrderUsecase) CreateOrder(ctx context.Context, userID uint64, reqItems []*CreateOrderRequestItem, addressID uint64, remark string) (*Order, error) {
	// TODO: 实现更复杂的订单创建逻辑，包括库存锁定、价格计算、优惠券应用等
	// 1. 获取 SKU 信息并计算总价
	// 2. 锁定库存 (分布式事务)
	// 3. 创建订单记录
	// 4. 清空购物车

	// 示例简化逻辑
	order := &Order{
		ID:          snowflake.GenID(),
		UserID:      userID,
		TotalAmount: 100.0, // 示例金额
		ShippingFee: uc.calculateShippingFee(), // 计算运费
		Status:      OrderStatusPending,
		AddressID:   addressID,
		Remark:      remark,
	}

	items := []*OrderItem{
		{SKUID: 1, ProductName: "示例商品", Price: 100.0, Quantity: 1},
	}

	createdOrder, err := uc.repo.CreateOrder(ctx, order, items)
	if err != nil {
		return nil, err
	}

	return createdOrder, nil
}

// GetPaymentURL 获取订单的支付链接
func (uc *OrderUsecase) GetPaymentURL(ctx context.Context, userID, orderID uint64) (string, error) {
	order, err := uc.repo.GetOrderByID(ctx, orderID)
	if err != nil {
		return "", fmt.Errorf("failed to get order: %w", err)
	}

	if order.UserID != userID {
		return "", ErrOrderBelongsToAnotherUser
	}

	if order.Status != OrderStatusPending {
		return "", ErrOrderNotPending
	}

	// 调用支付服务创建支付请求
	paymentInfo, err := uc.paymentClient.CreatePayment(ctx, userID, fmt.Sprintf("%d", orderID), order.TotalAmount)
	if err != nil {
		return "", fmt.Errorf("failed to create payment: %w", err)
	}

	return paymentInfo.PaymentURL, nil
}

// ProcessPaymentNotification 处理支付回调通知
func (uc *OrderUsecase) ProcessPaymentNotification(ctx context.Context, orderID uint64, paymentStatus paymentv1.PaymentStatus) error {
	order, err := uc.repo.GetOrderByID(ctx, orderID)
	if err != nil {
		return fmt.Errorf("failed to get order: %w", err)
	}

	if order.Status != OrderStatusPending {
		uc.logger.Warnf("Order %d is not pending, current status: %s", orderID, order.Status.String())
		return ErrOrderNotPending
	}

	var newOrderStatus OrderStatus
	switch paymentStatus {
	case paymentv1.PaymentStatus_PAYMENT_STATUS_SUCCESS:
		newOrderStatus = OrderStatusPaid
	case paymentv1.PaymentStatus_PAYMENT_STATUS_FAILED:
		newOrderStatus = OrderStatusCancelled // 支付失败，取消订单
	case paymentv1.PaymentStatus_PAYMENT_STATUS_CLOSED:
		newOrderStatus = OrderStatusCancelled // 交易关闭，取消订单
	default:
		return fmt.Errorf("unsupported payment status: %s", paymentStatus.String())
	}

	if err := uc.repo.UpdateOrderStatus(ctx, orderID, newOrderStatus); err != nil {
		return fmt.Errorf("failed to update order status: %w", err)
	}

	uc.logger.Infof("Order %d status updated to %s due to payment notification", orderID, newOrderStatus.String())
	return nil
}

// calculateShippingFee 计算运费 (TODO: 实现更复杂的运费计算逻辑)
func (uc *OrderUsecase) calculateShippingFee() float64 {
	return 10.0 // 示例：固定运费 10 元
}
