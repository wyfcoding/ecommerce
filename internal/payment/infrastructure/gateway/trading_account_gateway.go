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
	logging.Info(ctx, "trading gateway: Capture", "transaction_id", transactionID)

	// 由于 Capture 只传 ID，我们依赖 TccConfirmFreeze 的幂等性，且 OrderID 即为 transactionID
	resp, err := g.client.TccConfirmFreeze(ctx, &accountv1.TccFreezeRequest{
		OrderId: transactionID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to confirm trading account freeze: %w", err)
	}
	if !resp.Success {
		return nil, fmt.Errorf("trading account capture rejected: %s", resp.Message)
	}

	return &domain.PaymentGatewayResponse{TransactionID: transactionID}, nil
}

func (g *TradingAccountGateway) Void(ctx context.Context, transactionID string) error {
	logging.Info(ctx, "trading gateway: Void", "transaction_id", transactionID)
	resp, err := g.client.TccCancelFreeze(ctx, &accountv1.TccFreezeRequest{
		OrderId: transactionID,
	})
	if err != nil {
		return fmt.Errorf("failed to cancel trading account freeze: %w", err)
	}
	if !resp.Success {
		return fmt.Errorf("trading account void rejected: %s", resp.Message)
	}
	return nil
}

func (g *TradingAccountGateway) Refund(ctx context.Context, transactionID string, amount int64) error {
	// 退款逻辑：由于已经 Capture 成真钱了，这里可能需要调用原路退回逻辑（SagaRefundFrozen）
	logging.Info(ctx, "trading gateway: Refund", "transaction_id", transactionID, "amount", amount)
	return nil
}

func (g *TradingAccountGateway) DownloadBill(ctx context.Context, date time.Time) ([]*domain.GatewayBillItem, error) {
	return nil, nil
}
