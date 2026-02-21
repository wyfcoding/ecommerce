package domain

import (
	"time"

	pb "github.com/wyfcoding/ecommerce/go-api/order/v1"
)

// OrderCreatedEvent 订单创建事件
type OrderCreatedEvent struct {
	OrderID        uint64            `json:"order_id"`
	OrderNo        string            `json:"order_no"`
	UserID         uint64            `json:"user_id"`
	TotalAmount    int64             `json:"total_amount"`
	Status         pb.OrderStatus    `json:"status"`
	PaymentStatus  pb.PaymentStatus  `json:"payment_status"`
	ShippingStatus pb.ShippingStatus `json:"shipping_status"`
	OrderType      pb.OrderType      `json:"order_type"`
	DepositAmount  int64             `json:"deposit_amount"`
	BalanceAmount  int64             `json:"balance_amount"`
	Timestamp      time.Time         `json:"timestamp"`
}

func (e *OrderCreatedEvent) EventName() string { return "order.created" }
func (e *OrderCreatedEvent) EventKey() string  { return e.OrderNo }

// OrderPaidEvent 订单支付事件
type OrderPaidEvent struct {
	OrderID              uint64           `json:"order_id"`
	OrderNo              string           `json:"order_no"`
	UserID               uint64           `json:"user_id"`
	ActualAmount         int64            `json:"actual_amount"`
	PaymentMethod        string           `json:"payment_method"`
	PaymentTransactionID string           `json:"payment_transaction_id"`
	PaymentStatus        pb.PaymentStatus `json:"payment_status"`
	PaidAt               time.Time        `json:"paid_at"`
	Timestamp            time.Time        `json:"timestamp"`
}

func (e *OrderPaidEvent) EventName() string { return "order.paid" }
func (e *OrderPaidEvent) EventKey() string  { return e.OrderNo }

// OrderShippedEvent 订单发货事件
type OrderShippedEvent struct {
	OrderID          uint64            `json:"order_id"`
	OrderNo          string            `json:"order_no"`
	UserID           uint64            `json:"user_id"`
	TrackingNumber   string            `json:"tracking_number"`
	LogisticsCompany string            `json:"logistics_company"`
	ShippingStatus   pb.ShippingStatus `json:"shipping_status"`
	ShippedAt        time.Time         `json:"shipped_at"`
	Timestamp        time.Time         `json:"timestamp"`
}

func (e *OrderShippedEvent) EventName() string { return "order.shipped" }
func (e *OrderShippedEvent) EventKey() string  { return e.OrderNo }

// OrderDeliveredEvent 订单送达事件
type OrderDeliveredEvent struct {
	OrderID     uint64    `json:"order_id"`
	OrderNo     string    `json:"order_no"`
	UserID      uint64    `json:"user_id"`
	DeliveredAt time.Time `json:"delivered_at"`
	Timestamp   time.Time `json:"timestamp"`
}

func (e *OrderDeliveredEvent) EventName() string { return "order.delivered" }
func (e *OrderDeliveredEvent) EventKey() string  { return e.OrderNo }

// OrderPaymentStatusUpdatedEvent 订单支付状态更新事件
type OrderPaymentStatusUpdatedEvent struct {
	OrderID              uint64           `json:"order_id"`
	OrderNo              string           `json:"order_no"`
	UserID               uint64           `json:"user_id"`
	PaymentStatus        pb.PaymentStatus `json:"payment_status"`
	PaymentMethod        string           `json:"payment_method"`
	PaymentTransactionID string           `json:"payment_transaction_id"`
	Timestamp            time.Time        `json:"timestamp"`
}

func (e *OrderPaymentStatusUpdatedEvent) EventName() string { return "order.payment.status.updated" }
func (e *OrderPaymentStatusUpdatedEvent) EventKey() string  { return e.OrderNo }

// OrderShippingStatusUpdatedEvent 订单物流状态更新事件
type OrderShippingStatusUpdatedEvent struct {
	OrderID          uint64            `json:"order_id"`
	OrderNo          string            `json:"order_no"`
	UserID           uint64            `json:"user_id"`
	ShippingStatus   pb.ShippingStatus `json:"shipping_status"`
	TrackingNumber   string            `json:"tracking_number"`
	LogisticsCompany string            `json:"logistics_company"`
	Timestamp        time.Time         `json:"timestamp"`
}

func (e *OrderShippingStatusUpdatedEvent) EventName() string { return "order.shipping.updated" }
func (e *OrderShippingStatusUpdatedEvent) EventKey() string  { return e.OrderNo }

// OrderCompletedEvent 订单完成事件
type OrderCompletedEvent struct {
	OrderID     uint64    `json:"order_id"`
	OrderNo     string    `json:"order_no"`
	UserID      uint64    `json:"user_id"`
	CompletedAt time.Time `json:"completed_at"`
	Timestamp   time.Time `json:"timestamp"`
}

func (e *OrderCompletedEvent) EventName() string { return "order.completed" }
func (e *OrderCompletedEvent) EventKey() string  { return e.OrderNo }

// OrderCancelledEvent 订单取消事件
type OrderCancelledEvent struct {
	OrderID        uint64            `json:"order_id"`
	OrderNo        string            `json:"order_no"`
	UserID         uint64            `json:"user_id"`
	Reason         string            `json:"reason"`
	PaymentStatus  pb.PaymentStatus  `json:"payment_status"`
	ShippingStatus pb.ShippingStatus `json:"shipping_status"`
	CancelledAt    time.Time         `json:"cancelled_at"`
	Timestamp      time.Time         `json:"timestamp"`
}

func (e *OrderCancelledEvent) EventName() string { return "order.cancelled" }
func (e *OrderCancelledEvent) EventKey() string  { return e.OrderNo }

// OrderConfirmedEvent 订单确认事件 (Saga 确认后)
type OrderConfirmedEvent struct {
	OrderID   uint64    `json:"order_id"`
	OrderNo   string    `json:"order_no"`
	UserID    uint64    `json:"user_id"`
	Timestamp time.Time `json:"timestamp"`
}

func (e *OrderConfirmedEvent) EventName() string { return "order.confirmed" }
func (e *OrderConfirmedEvent) EventKey() string  { return e.OrderNo }

// OrderPaymentTimeoutEvent 订单支付超时预警事件
type OrderPaymentTimeoutEvent struct {
	OrderID   uint64    `json:"order_id"`
	OrderNo   string    `json:"order_no"`
	UserID    uint64    `json:"user_id"`
	ExpiresAt int64     `json:"expires_at"`
	Timestamp time.Time `json:"timestamp"`
}

func (e *OrderPaymentTimeoutEvent) EventName() string { return "order.payment.timeout" }
func (e *OrderPaymentTimeoutEvent) EventKey() string  { return e.OrderNo }

// OrderRefundRequestedEvent 订单退款申请事件
type OrderRefundRequestedEvent struct {
	OrderID       uint64           `json:"order_id"`
	OrderNo       string           `json:"order_no"`
	UserID        uint64           `json:"user_id"`
	RefundAmount  int64            `json:"refund_amount"`
	RefundReason  string           `json:"refund_reason"`
	PaymentStatus pb.PaymentStatus `json:"payment_status"`
	Timestamp     time.Time        `json:"timestamp"`
}

func (e *OrderRefundRequestedEvent) EventName() string { return "order.refund.requested" }
func (e *OrderRefundRequestedEvent) EventKey() string  { return e.OrderNo }

// OrderRefundApprovedEvent 订单退款完成事件
type OrderRefundApprovedEvent struct {
	OrderID       uint64           `json:"order_id"`
	OrderNo       string           `json:"order_no"`
	UserID        uint64           `json:"user_id"`
	RefundAmount  int64            `json:"refund_amount"`
	PaymentStatus pb.PaymentStatus `json:"payment_status"`
	Timestamp     time.Time        `json:"timestamp"`
}

func (e *OrderRefundApprovedEvent) EventName() string { return "order.refund.approved" }
func (e *OrderRefundApprovedEvent) EventKey() string  { return e.OrderNo }
