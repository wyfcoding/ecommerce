package domain

import (
	"time"

	pb "github.com/wyfcoding/ecommerce/goapi/order/v1"
)

// OrderCreatedEvent 订单创建事件
type OrderCreatedEvent struct {
	OrderID     uint64         `json:"order_id"`
	OrderNo     string         `json:"order_no"`
	UserID      uint64         `json:"user_id"`
	TotalAmount int64          `json:"total_amount"`
	Status      pb.OrderStatus `json:"status"`
	Timestamp   time.Time      `json:"timestamp"`
}

// OrderPaidEvent 订单支付事件
type OrderPaidEvent struct {
	OrderID              uint64    `json:"order_id"`
	OrderNo              string    `json:"order_no"`
	UserID               uint64    `json:"user_id"`
	ActualAmount         int64     `json:"actual_amount"`
	PaymentMethod        string    `json:"payment_method"`
	PaymentTransactionID string    `json:"payment_transaction_id"`
	PaidAt               time.Time `json:"paid_at"`
	Timestamp            time.Time `json:"timestamp"`
}

// OrderShippedEvent 订单发货事件
type OrderShippedEvent struct {
	OrderID          uint64    `json:"order_id"`
	OrderNo          string    `json:"order_no"`
	UserID           uint64    `json:"user_id"`
	TrackingNumber   string    `json:"tracking_number"`
	LogisticsCompany string    `json:"logistics_company"`
	ShippedAt        time.Time `json:"shipped_at"`
	Timestamp        time.Time `json:"timestamp"`
}

// OrderDeliveredEvent 订单送达事件
type OrderDeliveredEvent struct {
	OrderID     uint64    `json:"order_id"`
	OrderNo     string    `json:"order_no"`
	UserID      uint64    `json:"user_id"`
	DeliveredAt time.Time `json:"delivered_at"`
	Timestamp   time.Time `json:"timestamp"`
}

// OrderCompletedEvent 订单完成事件
type OrderCompletedEvent struct {
	OrderID     uint64    `json:"order_id"`
	OrderNo     string    `json:"order_no"`
	UserID      uint64    `json:"user_id"`
	CompletedAt time.Time `json:"completed_at"`
	Timestamp   time.Time `json:"timestamp"`
}

// OrderCancelledEvent 订单取消事件
type OrderCancelledEvent struct {
	OrderID     uint64    `json:"order_id"`
	OrderNo     string    `json:"order_no"`
	UserID      uint64    `json:"user_id"`
	Reason      string    `json:"reason"`
	CancelledAt time.Time `json:"cancelled_at"`
	Timestamp   time.Time `json:"timestamp"`
}

// OrderConfirmedEvent 订单确认事件 (Saga 确认后)
type OrderConfirmedEvent struct {
	OrderID   uint64    `json:"order_id"`
	OrderNo   string    `json:"order_no"`
	UserID    uint64    `json:"user_id"`
	Timestamp time.Time `json:"timestamp"`
}

// OrderPaymentTimeoutEvent 订单支付超时预警事件
type OrderPaymentTimeoutEvent struct {
	OrderID   uint64    `json:"order_id"`
	OrderNo   string    `json:"order_no"`
	UserID    uint64    `json:"user_id"`
	ExpiresAt int64     `json:"expires_at"`
	Timestamp time.Time `json:"timestamp"`
}

// OrderRefundRequestedEvent 订单退款申请事件
type OrderRefundRequestedEvent struct {
	OrderID      uint64    `json:"order_id"`
	OrderNo      string    `json:"order_no"`
	UserID       uint64    `json:"user_id"`
	RefundAmount int64     `json:"refund_amount"`
	RefundReason string    `json:"refund_reason"`
	Timestamp    time.Time `json:"timestamp"`
}

// OrderRefundApprovedEvent 订单退款完成事件
type OrderRefundApprovedEvent struct {
	OrderID   uint64    `json:"order_id"`
	OrderNo   string    `json:"order_no"`
	UserID    uint64    `json:"user_id"`
	Timestamp time.Time `json:"timestamp"`
}
