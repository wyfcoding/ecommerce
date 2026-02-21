package external

import (
	"context"

	"github.com/shopspring/decimal"
	"github.com/wyfcoding/ecommerce/internal/payment/domain"
	pb "github.com/wyfcoding/financialtrading/go-api/fiat/v1"
)

type fiatGrpcAdapter struct {
	client pb.FiatServiceClient
}

func NewFiatGrpcAdapter(client pb.FiatServiceClient) domain.FiatAdapter {
	return &fiatGrpcAdapter{client: client}
}

func (a *fiatGrpcAdapter) GetRate(ctx context.Context, from, to string) (decimal.Decimal, error) {
	resp, err := a.client.GetRate(ctx, &pb.GetRateRequest{
		FromCurrency: from,
		ToCurrency:   to,
	})
	if err != nil {
		return decimal.Zero, err
	}
	return decimal.NewFromFloat(resp.Rate), nil
}

func (a *fiatGrpcAdapter) LockRate(ctx context.Context, userID, paymentID, from, to string, amount decimal.Decimal) (string, decimal.Decimal, error) {
	amt, _ := amount.Float64()
	resp, err := a.client.LockRate(ctx, &pb.LockRateRequest{
		UserId:       userID,
		PaymentId:    paymentID,
		FromCurrency: from,
		ToCurrency:   to,
		Amount:       amt,
	})
	if err != nil {
		return "", decimal.Zero, err
	}
	return resp.LockId, decimal.NewFromFloat(resp.LockedRate), nil
}

func (a *fiatGrpcAdapter) VerifyLock(ctx context.Context, lockID string) (bool, decimal.Decimal, error) {
	resp, err := a.client.VerifyLock(ctx, &pb.VerifyLockRequest{
		LockId: lockID,
	})
	if err != nil {
		return false, decimal.Zero, err
	}
	return resp.Valid, decimal.NewFromFloat(resp.LockedRate), nil
}
