package application

import (
	"context"

	"github.com/wyfcoding/ecommerce/internal/payment/domain"
)

// PaymentQuery 支付查询服务。
type PaymentQuery struct {
	paymentRepo domain.PaymentRepository
}

// NewPaymentQuery 构造函数。
func NewPaymentQuery(paymentRepo domain.PaymentRepository) *PaymentQuery {
	return &PaymentQuery{
		paymentRepo: paymentRepo,
	}
}

// GetPayment 获取支付详情 (按 ID)。
func (q *PaymentQuery) GetPayment(ctx context.Context, userID uint64, paymentID uint64) (*domain.Payment, error) {
	return q.paymentRepo.FindByID(ctx, userID, paymentID)
}

// GetPaymentByNo 获取支付详情 (按支付单号)。
func (q *PaymentQuery) GetPaymentByNo(ctx context.Context, userID uint64, paymentNo string) (*domain.Payment, error) {
	return q.paymentRepo.FindByPaymentNo(ctx, userID, paymentNo)
}

// GetPaymentByOrder 获取订单支付信息。
func (q *PaymentQuery) GetPaymentByOrder(ctx context.Context, userID uint64, orderID uint64) (*domain.Payment, error) {
	return q.paymentRepo.FindByOrderID(ctx, userID, orderID)
}

// GetPaymentLogs 获取支付日志。
func (q *PaymentQuery) GetPaymentLogs(ctx context.Context, userID uint64, paymentID uint64) ([]*domain.PaymentLog, error) {
	return q.paymentRepo.FindLogsByPaymentID(ctx, userID, paymentID)
}

// GetUserIDByPaymentNo 根据支付单号查找用户 ID。
func (q *PaymentQuery) GetUserIDByPaymentNo(ctx context.Context, paymentNo string) (uint64, error) {
	return q.paymentRepo.GetUserIDByPaymentNo(ctx, paymentNo)
}

// GetPaymentStatus 获取支付状态 (按单号或 ID 的通用查询)。
func (q *PaymentQuery) GetPaymentStatus(ctx context.Context, userID uint64, identifier any) (*domain.Payment, error) {
	switch v := identifier.(type) {
	case string:
		return q.paymentRepo.FindByPaymentNo(ctx, userID, v)
	case uint64:
		return q.paymentRepo.FindByID(ctx, userID, v)
	default:
		return nil, domain.ErrInvalidParameter
	}
}
