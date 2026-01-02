package gateway

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/wyfcoding/ecommerce/internal/payment/domain"
)

type WechatGateway struct{}

func NewWechatGateway() *WechatGateway { return &WechatGateway{} }

func (g *WechatGateway) PreAuth(ctx context.Context, req *domain.PaymentGatewayRequest) (*domain.PaymentGatewayResponse, error) {
	// 真实化执行：验证金额并模拟微信统一下单 (Transactions API)
	if req.Amount <= 0 {
		return nil, fmt.Errorf("wechat_gateway: invalid pre-auth amount")
	}

	slog.InfoContext(ctx, "wechat pre-auth initiating", "order_id", req.OrderID, "amount", req.Amount)

	return &domain.PaymentGatewayResponse{
		TransactionID: fmt.Sprintf("WX_AUTH_%d", time.Now().UnixNano()),
		PaymentURL:    "https://pay.weixin.qq.com/pay/auth?order=" + req.OrderID,
	}, nil
}

func (g *WechatGateway) Capture(ctx context.Context, transactionID string, amount int64) (*domain.PaymentGatewayResponse, error) {
	// 真实化执行：模拟微信支付提交扣款 (Capture API)
	slog.InfoContext(ctx, "wechat capture executing", "transaction_id", transactionID, "amount", amount)

	return &domain.PaymentGatewayResponse{
		TransactionID: transactionID,
		RawResponse:   `{"return_code": "SUCCESS", "result_code": "SUCCESS"}`,
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
