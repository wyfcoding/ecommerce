// 变更说明：新增交易账户支付与资产质押支付聚合根，支持使用交易侧余额或持仓面值偿付订单。
// 假设：资产质押采用锁定持仓而非划转权属的方式，清算时由系统自动平仓。
package domain

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/shopspring/decimal"
	pb "github.com/wyfcoding/ecommerce/go-api/payment/v1"
)

// --- 交易账户支付聚合 ---

// TradingPayment 交易账户支付聚合根
type TradingPayment struct {
	ID               uint64           `json:"id"`
	CreatedAt        time.Time        `json:"created_at"`
	OrderID          uint64           `json:"order_id"`
	OrderNo          string           `json:"order_no"`
	UserID           string           `json:"user_id"`
	TradingAccountID string           `json:"trading_account_id"`
	Amount           int64            `json:"amount"`          // 订单总额（分）
	DiscountAmount   decimal.Decimal  `json:"discount_amount"` // 折扣金额
	ActualAmount     decimal.Decimal  `json:"actual_amount"`   // 实际应付金额
	Currency         string           `json:"currency"`
	Status           pb.PaymentStatus `json:"status"`
	TransactionID    string           `json:"transaction_id"` // 交易侧流水号
	FrozenAt         *time.Time       `json:"frozen_at"`
	DeductedAt       *time.Time       `json:"deducted_at"`
	LastError        string           `json:"last_error"`
}

// NewTradingPayment 创建交易账户支付任务
func NewTradingPayment(orderID uint64, orderNo, userID, accountID, currency string, amount int64) *TradingPayment {
	return &TradingPayment{
		OrderID:          orderID,
		OrderNo:          orderNo,
		UserID:           userID,
		TradingAccountID: accountID,
		Amount:           amount,
		Currency:         currency,
		Status:           pb.PaymentStatus_PENDING,
	}
}

// Intiate 执行预冻结
func (p *TradingPayment) Initiate(ctx context.Context, api FinancialSynergyService) error {
	if p.Status != pb.PaymentStatus_PENDING {
		return errors.New("invalid status for initiation")
	}

	req := &TradingPaymentRequest{
		OrderNo:     p.OrderNo,
		UserID:      p.UserID,
		AccountID:   p.TradingAccountID,
		Amount:      EcommerceToDecimal(p.Amount),
		Currency:    p.Currency,
		Description: fmt.Sprintf("Ecommerce Order %s Payment", p.OrderNo),
	}

	if err := api.FreezeTradingFunds(ctx, req); err != nil {
		p.LastError = err.Error()
		return err
	}

	now := time.Now()
	p.Status = pb.PaymentStatus_PENDING
	p.FrozenAt = &now
	return nil
}

// Confirm 确认并完成扣款
func (p *TradingPayment) Confirm(ctx context.Context, api FinancialSynergyService) error {
	if p.Status != pb.PaymentStatus_PENDING {
		return errors.New("must initiate before confirm")
	}

	req := &TradingPaymentRequest{
		OrderNo:   p.OrderNo,
		UserID:    p.UserID,
		AccountID: p.TradingAccountID,
		Amount:    EcommerceToDecimal(p.Amount),
		Currency:  p.Currency,
	}

	if err := api.DeductTradingFunds(ctx, req); err != nil {
		p.LastError = err.Error()
		return err
	}

	now := time.Now()
	p.Status = pb.PaymentStatus_SUCCESS
	p.DeductedAt = &now
	return nil
}

// ApplyVIPDiscount 应用交易账户 VIP 折扣
func (p *TradingPayment) ApplyVIPDiscount(vipLevel int32) {
	// 简单逻辑：每级 VIP 提供 1% 折扣，最高 10%
	discountRate := decimal.NewFromInt32(vipLevel).Mul(decimal.NewFromFloat(0.01))
	if discountRate.GreaterThan(decimal.NewFromFloat(0.1)) {
		discountRate = decimal.NewFromFloat(0.1)
	}

	p.DiscountAmount = EcommerceToDecimal(p.Amount).Mul(discountRate)
	p.ActualAmount = EcommerceToDecimal(p.Amount).Sub(p.DiscountAmount)
}

// --- 交易收益划转聚合 ---

// ProfitWithdrawal 交易获利提取聚合
type ProfitWithdrawal struct {
	WithdrawalID string
	UserID       string
	SourceSymbol string          // 获利来源标的
	ProfitAmount decimal.Decimal // 提取的利润金额
	TargetWallet string          // 目标电商钱包 ID
	Status       pb.PaymentStatus
	CreatedAt    time.Time
}

// NewProfitWithdrawal 创建收益提取请求
func NewProfitWithdrawal(userID, symbol string, amount decimal.Decimal) *ProfitWithdrawal {
	return &ProfitWithdrawal{
		WithdrawalID: fmt.Sprintf("PW-%d", time.Now().UnixNano()),
		UserID:       userID,
		SourceSymbol: symbol,
		ProfitAmount: amount,
		Status:       pb.PaymentStatus_PENDING,
		CreatedAt:    time.Now(),
	}
}

// --- 资产质押支付聚合 ---

// AssetCollateralPayment 资产质押支付聚合根
type AssetCollateralPayment struct {
	ID               uint64           `json:"id"`
	CreatedAt        time.Time        `json:"created_at"`
	OrderID          uint64           `json:"order_id"`
	UserID           string           `json:"user_id"`
	RequiredAmount   int64            `json:"required_amount"`    // 需偿付金额（分）
	PledgedAssets    []*PledgedAsset  `json:"pledged_assets"`     // 质押资产列表
	TotalCreditValue int64            `json:"total_credit_value"` // 总授信价值（分）
	Status           pb.PaymentStatus `json:"status"`
	LiquidatedAmount int64            `json:"liquidated_amount"` // 清算获偿金额
	CompletedAt      *time.Time       `json:"completed_at"`
}

// PledgedAsset 质押资产细则
type PledgedAsset struct {
	Symbol      string          `json:"symbol"`
	Quantity    decimal.Decimal `json:"quantity"`
	MarketPrice decimal.Decimal `json:"market_price"`
	Haircut     float64         `json:"haircut"`      // 折算率
	CreditValue int64           `json:"credit_value"` // 授信价值（分）
	Status      string          `json:"status"`       // PLEDGED/RELEASED/LIQUIDATED
}

// NewAssetCollateralPayment 创建质押支付任务
func NewAssetCollateralPayment(orderID uint64, userID string, requiredAmount int64) *AssetCollateralPayment {
	return &AssetCollateralPayment{
		OrderID:        orderID,
		UserID:         userID,
		RequiredAmount: requiredAmount,
		Status:         pb.PaymentStatus_PENDING,
		PledgedAssets:  make([]*PledgedAsset, 0),
	}
}

// AddCollateral 添加质押资产
func (p *AssetCollateralPayment) AddCollateral(symbol string, quantity decimal.Decimal, marketPrice decimal.Decimal, haircut float64) {
	creditDecimal := marketPrice.Mul(quantity).Mul(decimal.NewFromFloat(haircut))
	creditValue := DecimalToEcommerce(creditDecimal)

	p.PledgedAssets = append(p.PledgedAssets, &PledgedAsset{
		Symbol:      symbol,
		Quantity:    quantity,
		MarketPrice: marketPrice,
		Haircut:     haircut,
		CreditValue: creditValue,
		Status:      "PLEDGED",
	})
	p.TotalCreditValue += creditValue
}

// IsSufficient 检查质押价值是否覆盖订单金额
func (p *AssetCollateralPayment) IsSufficient() bool {
	return p.TotalCreditValue >= p.RequiredAmount
}

// ExecutePledge 执行交易侧持仓锁定
func (p *AssetCollateralPayment) ExecutePledge(ctx context.Context, api FinancialSynergyService) error {
	for _, asset := range p.PledgedAssets {
		if err := api.PledgeAsset(ctx, p.UserID, asset.Symbol, asset.Quantity, p.OrderID); err != nil {
			return err
		}
	}
	p.Status = pb.PaymentStatus_PENDING
	return nil
}

// Liquidate 执行强制平仓并清偿订单
func (p *AssetCollateralPayment) Liquidate(ctx context.Context, api FinancialSynergyService) error {
	if p.Status != pb.PaymentStatus_PENDING {
		return errors.New("can only liquidate after pledge is executed")
	}

	// 调用交易侧清算接口
	if err := api.LiquidateAndPay(ctx, p.OrderID, p.RequiredAmount); err != nil {
		return err
	}

	now := time.Now()
	p.Status = pb.PaymentStatus_SUCCESS
	p.LiquidatedAmount = p.RequiredAmount
	p.CompletedAt = &now

	for _, asset := range p.PledgedAssets {
		asset.Status = "LIQUIDATED"
	}

	return nil
}

// --- 仓储接口 ---

// SynergyPaymentRepository 协同支付仓储
type SynergyPaymentRepository interface {
	SaveTradingPayment(ctx context.Context, p *TradingPayment) error
	GetTradingPaymentByOrder(ctx context.Context, orderID uint64) (*TradingPayment, error)

	SaveAssetCollateral(ctx context.Context, p *AssetCollateralPayment) error
	GetAssetCollateralByOrder(ctx context.Context, orderID uint64) (*AssetCollateralPayment, error)
}
