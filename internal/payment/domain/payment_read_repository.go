// 生成摘要：定义支付读模型仓储接口（Redis），用于高频查询。
// 假设：读模型以支付ID/支付单号/订单ID为主键索引。
package domain

import "context"

// PaymentReadRepository 定义支付读模型的高性能访问接口。
type PaymentReadRepository interface {
	// Save 保存或更新支付读模型。
	Save(ctx context.Context, payment *Payment) error
	// GetByID 根据支付ID获取读模型。
	GetByID(ctx context.Context, userID uint64, paymentID uint64) (*Payment, error)
	// GetByPaymentNo 根据支付单号获取读模型。
	GetByPaymentNo(ctx context.Context, userID uint64, paymentNo string) (*Payment, error)
	// GetByOrderID 根据订单ID获取读模型。
	GetByOrderID(ctx context.Context, userID uint64, orderID uint64) (*Payment, error)
	// Delete 删除读模型数据（用于清理）。
	Delete(ctx context.Context, userID uint64, paymentID uint64, paymentNo string, orderID uint64) error
}
