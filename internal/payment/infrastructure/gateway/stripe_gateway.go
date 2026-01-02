package gateway

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/wyfcoding/ecommerce/internal/payment/domain"
)

type StripeGateway struct{}

func NewStripeGateway() *StripeGateway { return &StripeGateway{} }

func (g *StripeGateway) PreAuth(ctx context.Context, req *domain.PaymentGatewayRequest) (*domain.PaymentGatewayResponse, error) {
	// 真实化执行：验证金额并模拟 Stripe PaymentIntent
	if req.Amount < 50 {
		return nil, fmt.Errorf("stripe_gateway: amount below minimum threshold (50 cents)")
	}

	slog.InfoContext(ctx, "stripe payment_intent creating", "order_id", req.OrderID, "amount", req.Amount)

	return &domain.PaymentGatewayResponse{
		TransactionID: fmt.Sprintf("pi_%d_%s", time.Now().UnixNano(), req.OrderID),
		PaymentURL:    "https://checkout.stripe.com/pay/" + req.OrderID,
		RawResponse:   `{"status": "requires_payment_method"}`,
	}, nil
}

func (g *StripeGateway) Capture(ctx context.Context, transactionID string, amount int64) (*domain.PaymentGatewayResponse, error) {
	// 真实化执行：模拟支付确认
	slog.InfoContext(ctx, "stripe capture executing", "transaction_id", transactionID, "amount", amount)

	return &domain.PaymentGatewayResponse{
		TransactionID: transactionID,
		RawResponse:   `{"status": "succeeded"}`,
	}, nil
}

func (g *StripeGateway) Void(ctx context.Context, transactionID string) error {
	return nil
}

func (g *StripeGateway) Refund(ctx context.Context, transactionID string, amount int64) error {
	// 真实化执行：模拟调用 Stripe Refunds Create API
	if amount < 50 {
		return fmt.Errorf("stripe_gateway: refund amount below minimum (50 cents)")
	}
	slog.InfoContext(ctx, "stripe refund creating", "txn_id", transactionID, "amount", amount)
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
