package application

import (
	"context"
	"github.com/wyfcoding/ecommerce/internal/payment/domain"
	"github.com/wyfcoding/pkg/saga"
)

type PaymentAppService struct {
	repo domain.PaymentRepository
}

func (s *PaymentAppService) ExecutePaymentSaga(ctx context.Context, orderID string, amount int64) error {
	engine := saga.NewEngine()
	engine.AddStep("FreezeBalance",
		func(ctx context.Context) error { return nil },
		func(ctx context.Context) error { return nil },
	)
	return engine.Execute(ctx)
}
