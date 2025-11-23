package domain

import "context"

// PaymentRepository 支付仓储接口
type PaymentRepository interface {
	Save(ctx context.Context, payment *Payment) error
	FindByID(ctx context.Context, id uint64) (*Payment, error)
	FindByPaymentNo(ctx context.Context, paymentNo string) (*Payment, error)
	FindByOrderID(ctx context.Context, orderID uint64) (*Payment, error)
	Update(ctx context.Context, payment *Payment) error
	Delete(ctx context.Context, id uint64) error
	ListByUserID(ctx context.Context, userID uint64, offset, limit int) ([]*Payment, int64, error)
	SaveLog(ctx context.Context, log *PaymentLog) error
	FindLogsByPaymentID(ctx context.Context, paymentID uint64) ([]*PaymentLog, error)
}

// RefundRepository 退款仓储接口
type RefundRepository interface {
	Save(ctx context.Context, refund *Refund) error
	FindByID(ctx context.Context, id uint64) (*Refund, error)
	FindByRefundNo(ctx context.Context, refundNo string) (*Refund, error)
	Update(ctx context.Context, refund *Refund) error
	Delete(ctx context.Context, id uint64) error
}
