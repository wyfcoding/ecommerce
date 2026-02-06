package application

// InitiatePaymentCommand 发起支付命令。
type InitiatePaymentCommand struct {
	OrderID       uint64
	UserID        uint64
	Amount        int64
	PaymentMethod string
	ClientIP      string
	DeviceID      string
}

// CapturePaymentCommand 捕获支付命令。
type CapturePaymentCommand struct {
	UserID    uint64
	PaymentNo string
	Amount    int64
}

// RefundPaymentCommand 退款命令。
type RefundPaymentCommand struct {
	UserID    uint64
	PaymentID uint64
	Amount    int64
	Reason    string
}
