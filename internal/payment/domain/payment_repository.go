package domain

import "context"

// PaymentRepository 是支付模块的仓储接口。
// 它定义了对 Payment 和 PaymentLog 实体进行数据持久化操作的契约。
// 仓储接口属于领域层，旨在将领域逻辑与数据存储的实现细节解耦。
type PaymentRepository interface {
	// Save 将支付实体保存到数据存储中。
	// 如果支付已存在，则更新；如果不存在，则创建。
	// ctx: 上下文。
	// payment: 待保存的支付实体。
	Save(ctx context.Context, payment *Payment) error
	// FindByID 根据ID获取支付实体。
	FindByID(ctx context.Context, id uint64) (*Payment, error)
	// FindByPaymentNo 根据支付单号获取支付实体。
	FindByPaymentNo(ctx context.Context, paymentNo string) (*Payment, error)
	// FindByOrderID 根据订单ID获取支付实体。
	FindByOrderID(ctx context.Context, orderID uint64) (*Payment, error)
	// Update 更新支付实体。
	Update(ctx context.Context, payment *Payment) error
	// Delete 根据ID删除支付实体。
	Delete(ctx context.Context, id uint64) error
	// ListByUserID 列出指定用户ID的所有支付实体，支持分页。
	ListByUserID(ctx context.Context, userID uint64, offset, limit int) ([]*Payment, int64, error)
	// SaveLog 将支付日志实体保存到数据存储中。
	SaveLog(ctx context.Context, log *PaymentLog) error
	// FindLogsByPaymentID 根据支付ID获取所有支付日志实体。
	FindLogsByPaymentID(ctx context.Context, paymentID uint64) ([]*PaymentLog, error)
}

// RefundRepository 是退款模块的仓储接口。
// 它定义了对 Refund 实体进行数据持久化操作的契约。
// 仓储接口属于领域层，旨在将领域逻辑与数据存储的实现细节解耦。
type RefundRepository interface {
	// Save 将退款实体保存到数据存储中。
	// 如果退款已存在，则更新；如果不存在，则创建。
	// ctx: 上下文。
	// refund: 待保存的退款实体。
	Save(ctx context.Context, refund *Refund) error
	// FindByID 根据ID获取退款实体。
	FindByID(ctx context.Context, id uint64) (*Refund, error)
	// FindByRefundNo 根据退款单号获取退款实体。
	FindByRefundNo(ctx context.Context, refundNo string) (*Refund, error)
	// Update 更新退款实体。
	Update(ctx context.Context, refund *Refund) error
	// Delete 根据ID删除退款实体。
	Delete(ctx context.Context, id uint64) error
}
