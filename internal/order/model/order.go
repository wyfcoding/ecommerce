package model

import "time"

// OrderStatus 订单状态枚举
type OrderStatus int32

const (
	OrderStatusUnspecified OrderStatus = 0 // 未指定
	PendingPayment         OrderStatus = 1 // 待支付
	Paid                   OrderStatus = 2 // 已支付
	Shipped                OrderStatus = 3 // 已发货
	Delivered              OrderStatus = 4 // 已送达
	Completed              OrderStatus = 5 // 已完成
	Cancelled              OrderStatus = 6 // 已取消
	RefundRequested        OrderStatus = 7 // 退款申请中
	Refunded               OrderStatus = 8 // 已退款
	Closed                 OrderStatus = 9 // 已关闭
)

// PaymentStatus 支付状态枚举
type PaymentStatus int32

const (
	PaymentStatusUnspecified PaymentStatus = 0 // 未指定
	Unpaid                   PaymentStatus = 1 // 未支付
	Processing               PaymentStatus = 2 // 支付处理中
	Success                  PaymentStatus = 3 // 支付成功
	Failed                   PaymentStatus = 4 // 支付失败
	Refunding                PaymentStatus = 5 // 退款中
	RefundSuccess            PaymentStatus = 6 // 退款成功
	RefundFailed             PaymentStatus = 7 // 退款失败
)

// ShippingStatus 配送状态枚举
type ShippingStatus int32

const (
	ShippingStatusUnspecified ShippingStatus = 0 // 未指定
	PendingShipment           ShippingStatus = 1 // 待发货
	ShippingShipped           ShippingStatus = 2 // 已发货
	InTransit                 ShippingStatus = 3 // 运输中
	ShippingDelivered         ShippingStatus = 4 // 已送达
	Exception                 ShippingStatus = 5 // 配送异常
)

// Order 订单的业务领域模型（聚合根）
type Order struct {
	ID             uint64
	OrderNo        string
	UserID         uint64
	Status         OrderStatus
	PaymentStatus  PaymentStatus
	ShippingStatus ShippingStatus
	TotalAmount    int64
	ActualAmount   int64
	ShippingFee    int64
	DiscountAmount int64
	PaymentMethod  string
	Remark         string
	CreatedAt      time.Time
	UpdatedAt      time.Time
	PaidAt         *time.Time
	ShippedAt      *time.Time
	DeliveredAt    *time.Time
	CompletedAt    *time.Time
	CancelledAt    *time.Time

	// 聚合字段
	Items           []*OrderItem    // 订单项列表
	ShippingAddress ShippingAddress // 收货地址
	Logs            []*OrderLog     // 订单日志
}

// OrderItem 订单项的业务领域模型
type OrderItem struct {
	ID              uint64
	OrderID         uint64
	ProductID       uint64
	SkuID           uint64
	ProductName     string
	SkuName         string
	ProductImageURL string
	Price           int64
	Quantity        int32
	TotalPrice      int64
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// ShippingAddress 收货地址的业务领域模型
type ShippingAddress struct {
	ID              uint64
	OrderID         uint64
	RecipientName   string
	PhoneNumber     string
	Province        string
	City            string
	District        string
	DetailedAddress string
	PostalCode      string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// OrderLog 订单日志的业务领域模型
type OrderLog struct {
	ID        uint64
	OrderID   uint64
	Operator  string
	Action    string
	OldStatus string
	NewStatus string
	Remark    string
	CreatedAt time.Time
}

// String 方法实现
func (s OrderStatus) String() string {
	switch s {
	case PendingPayment:
		return "PendingPayment"
	case Paid:
		return "Paid"
	case Shipped:
		return "Shipped"
	case Delivered:
		return "Delivered"
	case Completed:
		return "Completed"
	case Cancelled:
		return "Cancelled"
	case RefundRequested:
		return "RefundRequested"
	case Refunded:
		return "Refunded"
	case Closed:
		return "Closed"
	default:
		return "Unspecified"
	}
}

func (s PaymentStatus) String() string {
	switch s {
	case Unpaid:
		return "Unpaid"
	case Processing:
		return "Processing"
	case Success:
		return "Success"
	case Failed:
		return "Failed"
	case Refunding:
		return "Refunding"
	case RefundSuccess:
		return "RefundSuccess"
	case RefundFailed:
		return "RefundFailed"
	default:
		return "Unspecified"
	}
}

func (s ShippingStatus) String() string {
	switch s {
	case PendingShipment:
		return "PendingShipment"
	case ShippingShipped:
		return "Shipped"
	case InTransit:
		return "InTransit"
	case ShippingDelivered:
		return "Delivered"
	case Exception:
		return "Exception"
	default:
		return "Unspecified"
	}
}
