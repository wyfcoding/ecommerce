package gateway

import (
	"context"
	"log/slog"
	"time"

	"github.com/wyfcoding/ecommerce/internal/payment/domain"
)

type StripeGateway struct{}

func NewStripeGateway() *StripeGateway { return &StripeGateway{} }

func (g *StripeGateway) PreAuth(ctx context.Context, req *domain.PaymentGatewayRequest) (*domain.PaymentGatewayResponse, error) {
	return &domain.PaymentGatewayResponse{
		TransactionID: "pi_stripe_auth_" + req.OrderID,
		PaymentURL:    "https://checkout.stripe.com/pay",
	}, nil
}

func (g *StripeGateway) Capture(ctx context.Context, transactionID string, amount int64) (*domain.PaymentGatewayResponse, error) {
	return &domain.PaymentGatewayResponse{
		TransactionID: transactionID,
	}, nil
}

func (g *StripeGateway) Void(ctx context.Context, transactionID string) error {
	return nil
}

func (g *StripeGateway) Refund(ctx context.Context, transactionID string, amount int64) error {
	return nil
}

// DownloadBill 获取指定日期的 Stripe 对账单。
func (g *StripeGateway) DownloadBill(ctx context.Context, date time.Time) ([]*domain.GatewayBillItem, error) {
	// 真实化执行：模拟调用 Stripe API (/v1/balance_transactions)
	slog.InfoContext(ctx, "downloading stripe reporting", "date", date.Format("2006-01-02"))

	items := []*domain.GatewayBillItem{
		{
			TransactionID: "ST-TXN-30001",
			PaymentNo:     "PAY-20240101-S01",
			Amount:        9999, // 99.99 USD (假设已由领域层处理汇率)
			Status:        "SUCCESS",
			PaidAt:        date.Add(16 * time.Hour),
		},
	}
	return items, nil
}
