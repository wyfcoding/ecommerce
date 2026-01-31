package gateway

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/wyfcoding/ecommerce/internal/payment/domain"
	accountv1 "github.com/wyfcoding/financialtrading/go-api/account/v1"
	"github.com/wyfcoding/pkg/logging"
)

// TradingAccountGateway 实现对接 FinancialTrading 证券账户的支付网关。
type TradingAccountGateway struct {
	client accountv1.AccountServiceClient
}

// NewTradingAccountGateway 创建交易账户网关实例。
func NewTradingAccountGateway(client accountv1.AccountServiceClient) *TradingAccountGateway {
	return &TradingAccountGateway{client: client}
}

// PreAuth 执行预授权（对应交易侧资金冻结）。
func (g *TradingAccountGateway) PreAuth(ctx context.Context, req *domain.PaymentGatewayRequest) (*domain.PaymentGatewayResponse, error) {
	logging.Info(ctx, "trading gateway: PreAuth", "order_id", req.OrderID, "user_id", req.UserID)

	// 调用交易侧 TCC TryFreeze
	resp, err := g.client.TccTryFreeze(ctx, &accountv1.TccFreezeRequest{
		UserId:   strconv.FormatUint(req.UserID, 10),
		Currency: req.Currency,
		Amount:   strconv.FormatFloat(float64(req.Amount)/100.0, 'f', 2, 64), // 假设电商单位是分，交易侧是元
		OrderId:  req.OrderID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to call trading account freeze: %w", err)
	}

	if !resp.Success {
		return nil, fmt.Errorf("trading account freeze rejected: %s", resp.Message)
	}

	return &domain.PaymentGatewayResponse{
		TransactionID: req.OrderID, // 使用订单号作为事务ID
		PaymentURL:    "internal://trading_account",
	}, nil
}

// Capture 执行扣款确认（对应交易侧确认扣款）。
func (g *TradingAccountGateway) Capture(ctx context.Context, transactionID string, amount int64) (*domain.PaymentGatewayResponse, error) {
	// 获取 UserID（实际场景应从 transactionID 或存储中逻辑关联，此处简化模拟）
	// 由于 Capture 接口通常只传 transactionID，我们可能需要在 Gateway 实现内部维护会话或依赖 context。
	// 这里演示调用 SagaDeductFrozen

	// 注意：此处需要 UserID，如果接口不支持，我们需要在 PreAuth 记录或通过特定规则解析。
	// 工业级实现通常会有一个中间转换层。

	logging.Info(ctx, "trading gateway: Capture", "transaction_id", transactionID)
	// 此处占位，实际逻辑应根据交易侧 TCC/Saga 语义调用 Confirm 或 Deduct
	return &domain.PaymentGatewayResponse{TransactionID: transactionID}, nil
}

func (g *TradingAccountGateway) Void(ctx context.Context, transactionID string) error {
	return nil
}

func (g *TradingAccountGateway) Refund(ctx context.Context, transactionID string, amount int64) error {
	return nil
}

func (g *TradingAccountGateway) DownloadBill(ctx context.Context, date time.Time) ([]*domain.GatewayBillItem, error) {
	return nil, nil
}
