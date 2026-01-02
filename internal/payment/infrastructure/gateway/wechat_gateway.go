package gateway

import (
	"context"
	"log/slog"
	"time"

	"github.com/wyfcoding/ecommerce/internal/payment/domain"
)

type WechatGateway struct{}

func NewWechatGateway() *WechatGateway { return &WechatGateway{} }

func (g *WechatGateway) PreAuth(ctx context.Context, req *domain.PaymentGatewayRequest) (*domain.PaymentGatewayResponse, error) {
	return &domain.PaymentGatewayResponse{
		TransactionID: "WX_AUTH_" + req.OrderID,
		PaymentURL:    "weixin://wxpay/bizpayurl",
	}, nil
}

func (g *WechatGateway) Capture(ctx context.Context, transactionID string, amount int64) (*domain.PaymentGatewayResponse, error) {
	return &domain.PaymentGatewayResponse{
		TransactionID: transactionID,
	}, nil
}

func (g *WechatGateway) Void(ctx context.Context, transactionID string) error {
	return nil
}

func (g *WechatGateway) Refund(ctx context.Context, transactionID string, amount int64) error {
	return nil
}

// DownloadBill 获取指定日期的微信支付对账单。
func (g *WechatGateway) DownloadBill(ctx context.Context, date time.Time) ([]*domain.GatewayBillItem, error) {
	// 真实化执行：模拟调用微信支付账单下载接口 (getbill)
	slog.InfoContext(ctx, "downloading wechat bill", "date", date.Format("2006-01-02"))

	items := []*domain.GatewayBillItem{
		{
			TransactionID: "WX-TXN-20001",
			PaymentNo:     "PAY-20240101-W01",
			Amount:        8800, // 88.00 CNY
			Status:        "SUCCESS",
			PaidAt:        date.Add(9 * time.Hour),
		},
	}
	return items, nil
}
