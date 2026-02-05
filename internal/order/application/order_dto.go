package application

import (
	"time"

	"github.com/wyfcoding/ecommerce/internal/order/domain"
)

type OrderDTO struct {
	ID              uint                `json:"id"`
	OrderNo         string              `json:"order_no"`
	UserID          uint64              `json:"user_id"`
	Status          string              `json:"status"`
	TotalAmount     int64               `json:"total_amount"`
	ActualAmount    int64               `json:"actual_amount"`
	ShippingFee     int64               `json:"shipping_fee"`
	DiscountAmount  int64               `json:"discount_amount"`
	PaymentMethod   string              `json:"payment_method"`
	Remark          string              `json:"remark"`
	ShippingAddress *ShippingAddressDTO `json:"shipping_address"`
	Items           []*OrderItemDTO     `json:"items"`
	Logs            []*OrderLogDTO      `json:"logs"`
	PaidAt          *time.Time          `json:"paid_at"`
	ShippedAt       *time.Time          `json:"shipped_at"`
	DeliveredAt     *time.Time          `json:"delivered_at"`
	CompletedAt     *time.Time          `json:"completed_at"`
	CancelledAt     *time.Time          `json:"cancelled_at"`
	CreatedAt       time.Time           `json:"created_at"`
	UpdatedAt       time.Time           `json:"updated_at"`
}

type OrderItemDTO struct {
	ID              uint   `json:"id"`
	ProductID       uint64 `json:"product_id"`
	SkuID           uint64 `json:"sku_id"`
	ProductName     string `json:"product_name"`
	SkuName         string `json:"sku_name"`
	ProductImageURL string `json:"product_image_url"`
	Price           int64  `json:"price"`
	Quantity        int32  `json:"quantity"`
	TotalPrice      int64  `json:"total_price"`
}

type ShippingAddressDTO struct {
	RecipientName   string  `json:"recipient_name"`
	PhoneNumber     string  `json:"phone_number"`
	Province        string  `json:"province"`
	City            string  `json:"city"`
	District        string  `json:"district"`
	DetailedAddress string  `json:"detailed_address"`
	PostalCode      string  `json:"postal_code"`
	Lat             float64 `json:"lat"`
	Lon             float64 `json:"lon"`
}

type OrderLogDTO struct {
	Operator  string    `json:"operator"`
	Action    string    `json:"action"`
	OldStatus string    `json:"old_status"`
	NewStatus string    `json:"new_status"`
	Remark    string    `json:"remark"`
	CreatedAt time.Time `json:"created_at"`
}

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
