package domain

import (
	"errors"
	"slices"
	"time"

	"github.com/shopspring/decimal"
)

type ChannelType string

const (
	ChannelDirectBank ChannelType = "DIRECT_BANK"
	ChannelShortcut   ChannelType = "SHORTCUT"
	ChannelWire       ChannelType = "WIRE"
	ChannelACH        ChannelType = "ACH"
	ChannelSEPA       ChannelType = "SEPA"
	ChannelSWIFT      ChannelType = "SWIFT"
)

type TransactionStatus string

const (
	TxStatusPending    TransactionStatus = "PENDING"
	TxStatusProcessing TransactionStatus = "PROCESSING"
	TxStatusSuccess    TransactionStatus = "SUCCESS"
	TxStatusFailed     TransactionStatus = "FAILED"
	TxStatusCancelled  TransactionStatus = "CANCELLED"
	TxStatusRefunded   TransactionStatus = "REFUNDED"
)

type TransactionType string

const (
	TxTypeDeposit  TransactionType = "DEPOSIT"
	TxTypeWithdraw TransactionType = "WITHDRAW"
	TxTypeTransfer TransactionType = "TRANSFER"
	TxTypeExchange TransactionType = "EXCHANGE"
)

type Currency struct {
	Code          string          `json:"code"`
	Symbol        string          `json:"symbol"`
	Name          string          `json:"name"`
	Precision     int32           `json:"precision"`
	IsActive      bool            `json:"is_active"`
	ExchangeRate  decimal.Decimal `json:"exchange_rate"`
	RateUpdatedAt *time.Time      `json:"rate_updated_at"`
}

type FiatTransaction struct {
	ID            uint64            `json:"id"`
	TransactionID string            `json:"transaction_id"`
	UserID        uint64            `json:"user_id"`
	Type          TransactionType   `json:"type"`
	Amount        decimal.Decimal   `json:"amount"`
	Currency      string            `json:"currency"`
	Channel       ChannelType       `json:"channel"`
	BankCode      string            `json:"bank_code"`
	BankAccountID uint64            `json:"bank_account_id"`
	Status        TransactionStatus `json:"status"`
	FeeAmount     decimal.Decimal   `json:"fee_amount"`
	FeeCurrency   string            `json:"fee_currency"`
	ExchangeRate  decimal.Decimal   `json:"exchange_rate"`
	ReferenceNo   string            `json:"reference_no"`
	ExternalTxID  string            `json:"external_tx_id"`
	FailReason    string            `json:"fail_reason"`
	ProcessedAt   *time.Time        `json:"processed_at"`
	CompletedAt   *time.Time        `json:"completed_at"`
	CreatedAt     time.Time         `json:"created_at"`
	UpdatedAt     time.Time         `json:"updated_at"`
	events        []DomainEvent
}

func NewFiatTransaction(transactionID string, userID uint64, txType TransactionType, amount decimal.Decimal, currency string, channel ChannelType) *FiatTransaction {
	return &FiatTransaction{
		TransactionID: transactionID,
		UserID:        userID,
		Type:          txType,
		Amount:        amount,
		Currency:      currency,
		Channel:       channel,
		Status:        TxStatusPending,
		FeeAmount:     decimal.Zero,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
		events:        make([]DomainEvent, 0),
	}
}

func (t *FiatTransaction) SetBankInfo(bankCode string, bankAccountID uint64) {
	t.BankCode = bankCode
	t.BankAccountID = bankAccountID
}

func (t *FiatTransaction) SetFee(amount decimal.Decimal, currency string) {
	t.FeeAmount = amount
	t.FeeCurrency = currency
}

func (t *FiatTransaction) SetExchangeRate(rate decimal.Decimal) {
	t.ExchangeRate = rate
}

func (t *FiatTransaction) SetReferenceNo(ref string) {
	t.ReferenceNo = ref
}

func (t *FiatTransaction) StartProcessing() error {
	if t.Status != TxStatusPending {
		return ErrInvalidTransactionStatus
	}
	t.Status = TxStatusProcessing
	t.UpdatedAt = time.Now()
	now := time.Now()
	t.ProcessedAt = &now
	t.addEvent(&TransactionProcessingEvent{
		TransactionID: t.TransactionID,
		UserID:        t.UserID,
		OccurredAt:    now,
	})
	return nil
}

func (t *FiatTransaction) Complete(externalTxID string) error {
	if t.Status != TxStatusProcessing {
		return ErrInvalidTransactionStatus
	}
	t.Status = TxStatusSuccess
	t.ExternalTxID = externalTxID
	now := time.Now()
	t.CompletedAt = &now
	t.UpdatedAt = now
	t.addEvent(&TransactionCompletedEvent{
		TransactionID: t.TransactionID,
		UserID:        t.UserID,
		Amount:        t.Amount,
		Currency:      t.Currency,
		ExternalTxID:  externalTxID,
		OccurredAt:    now,
	})
	return nil
}

func (t *FiatTransaction) Fail(reason string) error {
	t.Status = TxStatusFailed
	t.FailReason = reason
	t.UpdatedAt = time.Now()
	t.addEvent(&TransactionFailedEvent{
		TransactionID: t.TransactionID,
		UserID:        t.UserID,
		Reason:        reason,
		OccurredAt:    time.Now(),
	})
	return nil
}

func (t *FiatTransaction) Cancel(reason string) error {
	if t.Status == TxStatusSuccess || t.Status == TxStatusProcessing {
		return ErrInvalidTransactionStatus
	}
	t.Status = TxStatusCancelled
	t.FailReason = reason
	t.UpdatedAt = time.Now()
	t.addEvent(&TransactionCancelledEvent{
		TransactionID: t.TransactionID,
		UserID:        t.UserID,
		Reason:        reason,
		OccurredAt:    time.Now(),
	})
	return nil
}

func (t *FiatTransaction) Refund() error {
	if t.Status != TxStatusSuccess {
		return ErrInvalidTransactionStatus
	}
	t.Status = TxStatusRefunded
	t.UpdatedAt = time.Now()
	t.addEvent(&TransactionRefundedEvent{
		TransactionID: t.TransactionID,
		UserID:        t.UserID,
		Amount:        t.Amount,
		Currency:      t.Currency,
		OccurredAt:    time.Now(),
	})
	return nil
}

func (t *FiatTransaction) Events() []DomainEvent {
	return t.events
}

func (t *FiatTransaction) ClearEvents() {
	t.events = make([]DomainEvent, 0)
}

func (t *FiatTransaction) addEvent(event DomainEvent) {
	t.events = append(t.events, event)
}

type ExchangeRate struct {
	ID           uint64          `json:"id"`
	FromCurrency string          `json:"from_currency"`
	ToCurrency   string          `json:"to_currency"`
	Rate         decimal.Decimal `json:"rate"`
	BuyRate      decimal.Decimal `json:"buy_rate"`
	SellRate     decimal.Decimal `json:"sell_rate"`
	Source       string          `json:"source"`
	UpdatedAt    time.Time       `json:"updated_at"`
}

func NewExchangeRate(from, to string, rate, buyRate, sellRate decimal.Decimal, source string) *ExchangeRate {
	return &ExchangeRate{
		FromCurrency: from,
		ToCurrency:   to,
		Rate:         rate,
		BuyRate:      buyRate,
		SellRate:     sellRate,
		Source:       source,
		UpdatedAt:    time.Now(),
	}
}

func (r *ExchangeRate) Update(rate, buyRate, sellRate decimal.Decimal) {
	r.Rate = rate
	r.BuyRate = buyRate
	r.SellRate = sellRate
	r.UpdatedAt = time.Now()
}

type BankAccount struct {
	ID          uint64        `json:"id"`
	UserID      uint64        `json:"user_id"`
	BankName    string        `json:"bank_name"`
	BankCode    string        `json:"bank_code"`
	AccountName string        `json:"account_name"`
	AccountNo   string        `json:"account_no"`
	AccountType string        `json:"account_type"`
	Currency    string        `json:"currency"`
	Country     string        `json:"country"`
	SwiftCode   string        `json:"swift_code"`
	IBAN        string        `json:"iban"`
	IsVerified  bool          `json:"is_verified"`
	IsDefault   bool          `json:"is_default"`
	Status      AccountStatus `json:"status"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
}

type AccountStatus string

const (
	AccountStatusActive   AccountStatus = "ACTIVE"
	AccountStatusInactive AccountStatus = "INACTIVE"
	AccountStatusFrozen   AccountStatus = "FROZEN"
)

func NewBankAccount(userID uint64, bankName, bankCode, accountName, accountNo, currency string) *BankAccount {
	return &BankAccount{
		UserID:      userID,
		BankName:    bankName,
		BankCode:    bankCode,
		AccountName: accountName,
		AccountNo:   accountNo,
		Currency:    currency,
		IsVerified:  false,
		IsDefault:   false,
		Status:      AccountStatusActive,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

func (a *BankAccount) Verify() {
	a.IsVerified = true
	a.UpdatedAt = time.Now()
}

func (a *BankAccount) SetDefault(isDefault bool) {
	a.IsDefault = isDefault
	a.UpdatedAt = time.Now()
}

func (a *BankAccount) Freeze() {
	a.Status = AccountStatusFrozen
	a.UpdatedAt = time.Now()
}

func (a *BankAccount) Unfreeze() {
	a.Status = AccountStatusActive
	a.UpdatedAt = time.Now()
}

type FiatChannel struct {
	ID          uint64          `json:"id"`
	Code        string          `json:"code"`
	Name        string          `json:"name"`
	ChannelType ChannelType     `json:"channel_type"`
	Currencies  []string        `json:"currencies"`
	Countries   []string        `json:"countries"`
	MinAmount   decimal.Decimal `json:"min_amount"`
	MaxAmount   decimal.Decimal `json:"max_amount"`
	FeeRate     decimal.Decimal `json:"fee_rate"`
	FeeFixed    decimal.Decimal `json:"fee_fixed"`
	IsActive    bool            `json:"is_active"`
	Priority    int             `json:"priority"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

func (c *FiatChannel) SupportsCurrency(currency string) bool {
	return slices.Contains(c.Currencies, currency)
}

func (c *FiatChannel) SupportsCountry(country string) bool {
	return slices.Contains(c.Countries, country)
}

func (c *FiatChannel) CalculateFee(amount decimal.Decimal) decimal.Decimal {
	return amount.Mul(c.FeeRate).Add(c.FeeFixed)
}

var (
	ErrInvalidTransactionStatus = errors.New("invalid transaction status for this operation")
	ErrTransactionNotFound      = errors.New("transaction not found")
	ErrBankAccountNotFound      = errors.New("bank account not found")
	ErrCurrencyNotSupported     = errors.New("currency not supported")
	ErrExchangeRateNotFound     = errors.New("exchange rate not found")
	ErrChannelNotAvailable      = errors.New("payment channel not available")
	ErrAmountOutOfRange         = errors.New("amount out of range")
)
