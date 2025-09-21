package biz

import (
	"context"
	"encoding/json"
	"time"
)

// OrderStatus 定义订单状态常量。
const (
	OrderStatusPendingPayment = 1  // 待支付
	OrderStatusPaid           = 2  // 已支付
	OrderStatusShipped        = 3  // 已发货
	OrderStatusCompleted      = 4  // 已完成
	OrderStatusCancelled      = 5  // 已取消
	OrderStatusRefunded       = 6  // 已退款
)

// Order 是订单的业务领域模型。
type Order struct {
	ID              uint64
	UserID          uint64
	TotalAmount     uint64
	PaymentAmount   uint64
	ShippingFee     uint64
	Status          int8            // 订单状态，建议使用常量或枚举定义
	ShippingAddress json.RawMessage // 存储序列化后的地址信息
	CreatedAt       time.Time
}

// OrderItem 是订单商品的业务领域模型。
type OrderItem struct {
	ID           uint64
	OrderID      uint64
	SkuID        uint64
	SpuID        uint64
	ProductTitle string
	ProductImage string
	Price        uint64
	Quantity     uint32
	SubTotal     uint64
}

// SkuInfo 是订单服务内部使用的商品SKU信息DTO。
type SkuInfo struct {
	SkuID uint64
	SpuID uint64
	Price uint64
	Stock uint32
	Title string
	Image string
}

// Transaction 定义了事务管理器的接口。
type Transaction interface {
	// ExecTx 在一个事务中执行传入的函数。
	ExecTx(context.Context, func(ctx context.Context) error) error
}

// OrderRepo 定义了订单数据仓库的接口。
type OrderRepo interface {
	CreateOrder(ctx context.Context, order *Order) (*Order, error)
	CreateOrderItems(ctx context.Context, items []*OrderItem) error
	GetOrder(ctx context.Context, id uint64) (*Order, error)
	// ... 其他订单相关的数据库操作
}

// ProductClient 定义了订单服务依赖的商品服务客户端接口。
type ProductClient interface {
	GetSkuInfos(ctx context.Context, skuIDs []uint64) (map[uint64]*SkuInfo, error)
	LockStock(ctx context.Context, items map[uint64]uint32) error
	UnlockStock(ctx context.Context, items map[uint64]uint32) error
}

// CartClient 定义了订单服务依赖的购物车服务客户端接口。
type CartClient interface {
	ClearCheckedItems(ctx context.Context, userID uint64) error
}
