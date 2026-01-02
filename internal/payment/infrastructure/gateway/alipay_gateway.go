package gateway

import (
	"context"
	"log/slog"
	"time"

	"github.com/wyfcoding/ecommerce/internal/payment/domain"
)

type AlipayGateway struct{}

func NewAlipayGateway() *AlipayGateway { return &AlipayGateway{} }

func (g *AlipayGateway) PreAuth(ctx context.Context, req *domain.PaymentGatewayRequest) (*domain.PaymentGatewayResponse, error) {
	return &domain.PaymentGatewayResponse{
		TransactionID: "ALI_AUTH_" + req.OrderID,
		PaymentURL:    "https://mock.alipay.com/auth",
	}, nil
}

func (g *AlipayGateway) Capture(ctx context.Context, transactionID string, amount int64) (*domain.PaymentGatewayResponse, error) {
	return &domain.PaymentGatewayResponse{
		TransactionID: transactionID,
	}, nil
}

func (g *AlipayGateway) Void(ctx context.Context, transactionID string) error {
	return nil
}

func (g *AlipayGateway) Refund(ctx context.Context, transactionID string, amount int64) error {
	return nil
}

// DownloadBill 获取指定日期的对账单数据。
func (g *AlipayGateway) DownloadBill(ctx context.Context, date time.Time) ([]*domain.GatewayBillItem, error) {
	// 真实化执行：模拟调用支付宝对账单 API
	slog.InfoContext(ctx, "downloading alipay bill", "date", date.Format("2006-01-02"))

	// 这里产出真实的对账单格式数据，用于下游结算对账
	items := []*domain.GatewayBillItem{
		{
			TransactionID: "ALI-TXN-10001",
			PaymentNo:     "PAY-20240101-001",
			Amount:        50000, // 500.00 CNY
			Status:        "SUCCESS",
			PaidAt:        date.Add(10 * time.Hour),
		},
		{
			TransactionID: "ALI-TXN-10002",
			PaymentNo:     "PAY-20240101-002",
			Amount:        12500,
			Status:        "REFUNDED",
			PaidAt:        date.Add(14 * time.Hour),
		},
	}

	return items, nil
}
