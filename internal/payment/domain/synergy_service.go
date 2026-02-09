// 变更说明：新增跨项目协同领域服务，定义 Ecommerce 与 FinancialTrading 之间的业务协议模型，支持交易账户结算与资产质押。
// 假设：FinancialTrading 侧支持订单维度的资金冻结与异步扣款。
package domain

import (
	"context"
	"time"

	"github.com/shopspring/decimal"
)

// --- 交易账户支付协议 ---

// TradingAccount 交易账户摘要信息
type TradingAccount struct {
	UserID           string          `json:"user_id"`
	AccountID        string          `json:"account_id"`
	Currency         string          `json:"currency"`
	AvailableBalance decimal.Decimal `json:"available_balance"`
	Status           string          `json:"status"` // ACTIVE/FROZEN
}

// TradingPaymentRequest 交易账户支付请求
type TradingPaymentRequest struct {
	OrderNo     string          `json:"order_no"`
	UserID      string          `json:"user_id"`
	AccountID   string          `json:"account_id"`
	Amount      decimal.Decimal `json:"amount"`
	Currency    string          `json:"currency"`
	Description string          `json:"description"`
}

// --- 资产质押协议 ---

// AssetHolding 资产持仓摘要
type AssetHolding struct {
	Symbol      string          `json:"symbol"`
	Quantity    decimal.Decimal `json:"quantity"`
	MarketPrice decimal.Decimal `json:"market_price"`
	Value       decimal.Decimal `json:"value"`        // 原始市值
	Haircut     float64         `json:"haircut"`      // 折算率（如 0.5 表示折算50%）
	CreditValue decimal.Decimal `json:"credit_value"` // 授信价值 (Value * Haircut)
}

// CollateralPledge 质押记录
type CollateralPledge struct {
	OrderID   uint64          `json:"order_id"`
	UserID    string          `json:"user_id"`
	AssetID   string          `json:"asset_id"`
	Symbol    string          `json:"symbol"`
	Quantity  decimal.Decimal `json:"quantity"`
	PledgedAt time.Time       `json:"pledged_at"`
	Status    string          `json:"status"` // PLEDGED/RELEASED/LIQUIDATED
}

// --- 协同领域服务接口 ---

// FinancialSynergyService 跨项目协同领域服务接口
type FinancialSynergyService interface {
	// 获取用户交易账户信息
	GetTradingAccount(ctx context.Context, userID string) (*TradingAccount, error)

	// 交易账户资金操作
	FreezeTradingFunds(ctx context.Context, req *TradingPaymentRequest) error
	UnfreezeTradingFunds(ctx context.Context, req *TradingPaymentRequest) error
	DeductTradingFunds(ctx context.Context, req *TradingPaymentRequest) error

	// 资产持仓操作
	GetUserHoldings(ctx context.Context, userID string) ([]*AssetHolding, error)
	PledgeAsset(ctx context.Context, userID string, symbol string, quantity decimal.Decimal, orderID uint64) error
	ReleaseAsset(ctx context.Context, orderID uint64) error

	// 强制平仓清算（用于支付质押订单）
	LiquidateAndPay(ctx context.Context, orderID uint64, amount int64) error
}

// --- 内部适配：Ecommerce 金额转换 ---

// EcommerceToDecimal 将 Ecommerce 的分转为 Decimal
func EcommerceToDecimal(amount int64) decimal.Decimal {
	return decimal.NewFromInt(amount).Div(decimal.NewFromInt(100))
}

// DecimalToEcommerce 将 Decimal 转为 Ecommerce 的分
func DecimalToEcommerce(amount decimal.Decimal) int64 {
	return amount.Mul(decimal.NewFromInt(100)).IntPart()
}
