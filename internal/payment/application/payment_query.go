package application

import (
	"context"

	"github.com/wyfcoding/ecommerce/internal/payment/domain"
)

// PaymentQuery 负责处理 Payment 相关的读操作和查询逻辑。
type PaymentQuery struct {
	paymentRepo domain.PaymentRepository
}

// NewPaymentQuery 负责处理 NewPayment 相关的读操作和查询逻辑。
func NewPaymentQuery(paymentRepo domain.PaymentRepository) *PaymentQuery {
	return &PaymentQuery{
		paymentRepo: paymentRepo,
	}
}

// GetPaymentStatus 获取支付详情
func (s *PaymentQuery) GetPaymentStatus(ctx context.Context, id uint64) (*domain.Payment, error) {
	return s.paymentRepo.FindByID(ctx, id)
}
