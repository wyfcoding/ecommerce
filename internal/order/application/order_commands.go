package application

import (
	pb "github.com/wyfcoding/ecommerce/goapi/order/v1"
	"github.com/wyfcoding/ecommerce/internal/order/domain"
)

// Commands

type CreateOrderCommand struct {
	UserID          uint64
	Items           []*CreateOrderItemCommand
	ShippingAddress *domain.ShippingAddress
	CouponCode      string
	Remark          string
	PaymentMethod   string
	ClientIP        string
	DeviceID        string
}

type CreateOrderItemCommand struct {
	SkuID     uint64
	ProductID uint64
	Quantity  int32
	Price     int64 // Optional validation or logic
}

type PayOrderCommand struct {
	UserID        uint64
	OrderID       uint64
	PaymentMethod string
	Amount        int64
	TransactionID string
}

type UpdatePaymentStatusCommand struct {
	UserID        uint64
	OrderID       uint64
	Operator      string
	Status        pb.PaymentStatus
	PaymentMethod string
	TransactionID string
	Remark        string
}

type ShipOrderCommand struct {
	UserID           uint64
	OrderID          uint64
	Operator         string
	TrackingNumber   string
	LogisticsCompany string
}

type DeliverOrderCommand struct {
	UserID           uint64
	OrderID          uint64
	Operator         string
	TrackingNumber   string
	LogisticsCompany string
}

type UpdateShippingStatusCommand struct {
	UserID           uint64
	OrderID          uint64
	Operator         string
	NewStatus        pb.ShippingStatus
	TrackingNumber   string
	LogisticsCompany string
	Remark           string
}

type CompleteOrderCommand struct {
	UserID   uint64
	OrderID  uint64
	Operator string
}

type CancelOrderCommand struct {
	UserID   uint64
	OrderID  uint64
	Operator string
	Reason   string
}

type RequestRefundCommand struct {
	UserID       uint64
	OrderID      uint64
	Operator     string
	RefundAmount int64
	Reason       string
}

type ApproveRefundCommand struct {
	UserID   uint64
	OrderID  uint64
	Operator string
}
