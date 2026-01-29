package application

import (
	"time"

	"github.com/wyfcoding/ecommerce/internal/payment/domain"
)

type PaymentDTO struct {
	ID             uint                 `json:"id"`
	PaymentNo      string               `json:"payment_no"`
	OrderID        uint64               `json:"order_id"`
	OrderNo        string               `json:"order_no"`
	UserID         uint64               `json:"user_id"`
	Amount         int64                `json:"amount"`
	CapturedAmount int64                `json:"captured_amount"`
	Currency       string               `json:"currency"`
	PaymentMethod  string               `json:"payment_method"`
	GatewayType    domain.GatewayType   `json:"gateway_type"`
	Status         domain.PaymentStatus `json:"status"`
	TransactionID  string               `json:"transaction_id"`
	PaidAt         *time.Time           `json:"paid_at"`
	CreatedAt      time.Time            `json:"created_at"`
	UpdatedAt      time.Time            `json:"updated_at"`
}

type InitiatePaymentCommand struct {
	OrderID       uint64
	UserID        uint64
	Amount        int64
	PaymentMethod string
	ClientIP      string
	DeviceID      string
}

type CapturePaymentCommand struct {
	UserID    uint64
	PaymentNo string
	Amount    int64
}

type RefundPaymentCommand struct {
	UserID    uint64
	PaymentID uint64
	Amount    int64
	Reason    string
}
