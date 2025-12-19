package application

import (
	"context"

	"github.com/wyfcoding/ecommerce/internal/payment/domain"
)

type PaymentQuery struct {
	paymentRepo domain.PaymentRepository
}

func NewPaymentQuery(paymentRepo domain.PaymentRepository) *PaymentQuery {
	return &PaymentQuery{
		paymentRepo: paymentRepo,
	}
}

// GetPaymentStatus 获取支付详情
func (s *PaymentQuery) GetPaymentStatus(ctx context.Context, id uint64) (*domain.Payment, error) {
	return s.paymentRepo.FindByID(ctx, id)
}
