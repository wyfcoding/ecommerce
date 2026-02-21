package domain

import (
	"context"
	"fmt"
	"time"

	"github.com/shopspring/decimal"
)

// CurrencyPair 货币对
type CurrencyPair struct {
	BaseCurrency  string `json:"base_currency"`
	QuoteCurrency string `json:"quote_currency"`
	Symbol        string `json:"symbol"` // 例如: USD/CNY
}

// ExchangeRate 汇率
type ExchangeRate struct {
	ID           string    `json:"id"`
	CurrencyPair string    `json:"currency_pair"`
	Rate         float64   `json:"rate"`
	Bid          float64   `json:"bid"`
	Ask          float64   `json:"ask"`
	Mid          float64   `json:"mid"`
	Spread       float64   `json:"spread"`
	Source       string    `json:"source"`
	Timestamp    time.Time `json:"timestamp"`
	ValidUntil   time.Time `json:"valid_until"`
	CreatedAt    time.Time `json:"created_at"`
}

// RateLock 汇率锁定
type RateLock struct {
	ID           string  `json:"id"`
	LockID       string  `json:"lock_id"`
	UserID       uint64  `json:"user_id"`
	PaymentID    uint    `json:"payment_id"`
	CurrencyPair string  `json:"currency_pair"`
	LockedRate   float64 `json:"locked_rate"`
	Amount       float64 `json:"amount"`
}

// ExchangeRateConfig 汇率配置
type ExchangeRateConfig struct {
	FXFeeRate float64 `json:"fx_fee_rate"`
	MinFXFee  float64 `json:"min_fx_fee"`
}

// CrossBorderPayment 跨境支付
type CrossBorderPayment struct {
	ID             string    `json:"id"`
	PaymentID      uint      `json:"payment_id"`
	UserID         uint64    `json:"user_id"`
	SourceCurrency string    `json:"source_currency"`
	TargetCurrency string    `json:"target_currency"`
	SourceAmount   int64     `json:"source_amount"`
	TargetAmount   float64   `json:"target_amount"`
	ExchangeRate   float64   `json:"exchange_rate"`
	RateLockID     string    `json:"rate_lock_id"`
	FXFee          float64   `json:"fx_fee"`
	TotalCost      float64   `json:"total_cost"`
	Status         string    `json:"status"` // PENDING, COMPLETED, FAILED
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// ExchangeRateService 汇率服务
type ExchangeRateService struct {
	fiatAdapter FiatAdapter
	config      *ExchangeRateConfig
}

// NewExchangeRateService 创建汇率服务
func NewExchangeRateService(fiatAdapter FiatAdapter) *ExchangeRateService {
	return &ExchangeRateService{
		fiatAdapter: fiatAdapter,
		config: &ExchangeRateConfig{
			FXFeeRate: 0.01,
			MinFXFee:  1,
		},
	}
}

// GetExchangeRate 获取当前汇率
func (ers *ExchangeRateService) GetExchangeRate(ctx context.Context, fromCurrency, toCurrency string) (decimal.Decimal, error) {
	return ers.fiatAdapter.GetRate(ctx, fromCurrency, toCurrency)
}

// CreateCrossBorderPayment 创建跨境支付
func (ers *ExchangeRateService) CreateCrossBorderPayment(ctx context.Context, payment *Payment, rateLockID string) (*CrossBorderPayment, error) {
	var exchangeRate decimal.Decimal
	var err error

	if rateLockID != "" {
		valid, rate, err := ers.fiatAdapter.VerifyLock(ctx, rateLockID)
		if err != nil || !valid {
			return nil, fmt.Errorf("invalid or expired rate lock: %w", err)
		}
		exchangeRate = rate
	} else {
		exchangeRate, err = ers.fiatAdapter.GetRate(ctx, payment.Currency, "CNY")
		if err != nil {
			return nil, fmt.Errorf("failed to get current exchange rate: %w", err)
		}
	}

	// 计算目标金额 (CNY)
	paymentAmt := decimal.NewFromInt(payment.Amount)
	targetAmount := paymentAmt.Mul(exchangeRate)

	// 计算外汇费用
	fxFee := paymentAmt.Mul(decimal.NewFromFloat(ers.config.FXFeeRate))
	if fxFee.LessThan(decimal.NewFromFloat(ers.config.MinFXFee)) {
		fxFee = decimal.NewFromFloat(ers.config.MinFXFee)
	}

	fxFeeFloat, _ := fxFee.Float64()
	targetAmtFloat, _ := targetAmount.Float64()
	rateFloat, _ := exchangeRate.Float64()

	crossBorderPayment := &CrossBorderPayment{
		ID:             generateCrossBorderID(),
		PaymentID:      payment.ID,
		UserID:         payment.UserID,
		SourceCurrency: payment.Currency,
		TargetCurrency: "CNY",
		SourceAmount:   payment.Amount,
		TargetAmount:   targetAmtFloat,
		ExchangeRate:   rateFloat,
		RateLockID:     rateLockID,
		FXFee:          fxFeeFloat,
		TotalCost:      float64(payment.Amount) + fxFeeFloat,
		Status:         "PENDING",
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	return crossBorderPayment, nil
}

// ProcessCrossBorderPayment 处理跨境支付
func (ers *ExchangeRateService) ProcessCrossBorderPayment(ctx context.Context, crossBorderPayment *CrossBorderPayment) error {
	// 检查状态
	if crossBorderPayment.Status != "PENDING" {
		return fmt.Errorf("cross border payment is not pending: %s", crossBorderPayment.Status)
	}

	// 执行外汇转换
	err := ers.executeFXConversion(ctx, crossBorderPayment)
	if err != nil {
		crossBorderPayment.Status = "FAILED"
		crossBorderPayment.UpdatedAt = time.Now()
		return fmt.Errorf("failed to execute FX conversion: %w", err)
	}

	// 更新状态
	crossBorderPayment.Status = "COMPLETED"
	crossBorderPayment.UpdatedAt = time.Now()

	return nil
}

// executeFXConversion 执行外汇转换
func (ers *ExchangeRateService) executeFXConversion(ctx context.Context, crossBorderPayment *CrossBorderPayment) error {
	// 实际应该调用银行或支付网关的API
	return nil
}

// Helper functions (Simplified)
func generateCrossBorderID() string {
	return fmt.Sprintf("CROSSBORDER_%d", time.Now().UnixNano())
}

// Repository interfaces

type ExchangeRateRepository interface {
	SaveExchangeRate(ctx context.Context, rate *ExchangeRate) error
	GetExchangeRate(ctx context.Context, currencyPair string) (*ExchangeRate, error)
	GetLatestExchangeRate(ctx context.Context, currencyPair string) (*ExchangeRate, error)
	GetExchangeRateHistory(ctx context.Context, currencyPair string, startDate, endDate time.Time) ([]*ExchangeRate, error)
	UpdateExchangeRate(ctx context.Context, rate *ExchangeRate) error
	DeleteExchangeRate(ctx context.Context, id string) error
}

type RateLockRepository interface {
	SaveRateLock(ctx context.Context, rateLock *RateLock) error
	GetRateLock(ctx context.Context, lockID string) (*RateLock, error)
	GetRateLockByPayment(ctx context.Context, paymentID string) (*RateLock, error)
	GetActiveRateLocksByUser(ctx context.Context, userID uint64) ([]*RateLock, error)
	GetExpiredRateLocks(ctx context.Context) ([]*RateLock, error)
	UpdateRateLock(ctx context.Context, rateLock *RateLock) error
	DeleteRateLock(ctx context.Context, lockID string) error
}
