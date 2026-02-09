// Package domain 提供了法币通道服务的领域模型。
// 变更说明：实现法币支付通道（Fiat）基础逻辑，支持银行卡直连、快捷支付与多币种汇率转换。
package domain

import (
	"context"
	"errors"
	"time"

	"github.com/shopspring/decimal"
)

// ChannelType 支付通道类型
type ChannelType string

const (
	ChannelDirectBank ChannelType = "DIRECT_BANK" // 银行直连
	ChannelShortcut   ChannelType = "SHORTCUT"    // 快捷支付
	ChannelWire       ChannelType = "WIRE"        // 跨境汇款
)

// Currency 币种模型
type Currency struct {
	Code      string // e.g., USD, CNY, EUR
	Symbol    string // e.g., $, ￥, €
	Precision int32
}

// FiatTransaction 法币交易记录
type FiatTransaction struct {
	TransactionID string
	UserID        string
	Amount        decimal.Decimal
	Currency      string
	Channel       ChannelType
	BankCode      string
	Status        string // PENDING, SUCCESS, FAILED
	CreatedAt     time.Time
}

// ExchangeRate 汇率模型
type ExchangeRate struct {
	FromCurrency string
	ToCurrency   string
	Rate         decimal.Decimal
	UpdatedAt    time.Time
}

// FiatRepository 法币通道仓储接口
type FiatRepository interface {
	SaveTransaction(ctx context.Context, tx *FiatTransaction) error
	GetTransaction(ctx context.Context, txID string) (*FiatTransaction, error)
	GetExchangeRate(ctx context.Context, from, to string) (*ExchangeRate, error)
}

// FiatPaymentService 法币支付通道服务
type FiatPaymentService interface {
	// InitiateDeposit 发起充值/付款
	InitiateDeposit(ctx context.Context, userID string, amount decimal.Decimal, currency string, channel ChannelType) (string, error)
	// ProcessCallback 处理银行回调
	ProcessCallback(ctx context.Context, transactionID string, success bool, rawData string) error
	// ConvertCurrency 币种转换
	ConvertCurrency(ctx context.Context, amount decimal.Decimal, from, to string) (decimal.Decimal, error)
}

// FiatCalculator 内部汇率计算器
type FiatCalculator struct {
	repo FiatRepository
}

func NewFiatCalculator(repo FiatRepository) *FiatCalculator {
	return &FiatCalculator{repo: repo}
}

// Calculate 把金额转为目标币种
func (c *FiatCalculator) Calculate(ctx context.Context, amount decimal.Decimal, from, to string) (decimal.Decimal, error) {
	if from == to {
		return amount, nil
	}
	rate, err := c.repo.GetExchangeRate(ctx, from, to)
	if err != nil {
		return decimal.Zero, err
	}
	if rate.Rate.IsZero() {
		return decimal.Zero, errors.New("invalid exchange rate")
	}
	return amount.Mul(rate.Rate), nil
}
