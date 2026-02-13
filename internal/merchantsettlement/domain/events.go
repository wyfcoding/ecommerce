package domain

import (
	"time"

	"github.com/shopspring/decimal"
)

type DomainEvent interface {
	EventType() string
	OccurredAt() time.Time
	AggregateID() string
}

type SettlementCreatedEvent struct {
	SettlementID string          `json:"settlement_id"`
	MerchantID   uint64          `json:"merchant_id"`
	Cycle        SettlementCycle `json:"cycle"`
	PeriodStart  time.Time       `json:"period_start"`
	PeriodEnd    time.Time       `json:"period_end"`
	OccurredAt   time.Time       `json:"occurred_at"`
}

func (e *SettlementCreatedEvent) EventType() string    { return "settlement.created" }
func (e *SettlementCreatedEvent) AggregateID() string  { return e.SettlementID }
func (e *SettlementCreatedEvent) OccurredAt() time.Time { return e.OccurredAt }

type SettlementCalculationStartedEvent struct {
	SettlementID string    `json:"settlement_id"`
	MerchantID   uint64    `json:"merchant_id"`
	OccurredAt   time.Time `json:"occurred_at"`
}

func (e *SettlementCalculationStartedEvent) EventType() string    { return "settlement.calculation_started" }
func (e *SettlementCalculationStartedEvent) AggregateID() string  { return e.SettlementID }
func (e *SettlementCalculationStartedEvent) OccurredAt() time.Time { return e.OccurredAt }

type SettlementCalculationCompletedEvent struct {
	SettlementID     string          `json:"settlement_id"`
	MerchantID       uint64          `json:"merchant_id"`
	SettlementAmount decimal.Decimal `json:"settlement_amount"`
	OrderCount       int64           `json:"order_count"`
	OccurredAt       time.Time       `json:"occurred_at"`
}

func (e *SettlementCalculationCompletedEvent) EventType() string    { return "settlement.calculation_completed" }
func (e *SettlementCalculationCompletedEvent) AggregateID() string  { return e.SettlementID }
func (e *SettlementCalculationCompletedEvent) OccurredAt() time.Time { return e.OccurredAt }

type SettlementApprovedEvent struct {
	SettlementID string    `json:"settlement_id"`
	MerchantID   uint64    `json:"merchant_id"`
	ApprovedBy   uint64    `json:"approved_by"`
	OccurredAt   time.Time `json:"occurred_at"`
}

func (e *SettlementApprovedEvent) EventType() string    { return "settlement.approved" }
func (e *SettlementApprovedEvent) AggregateID() string  { return e.SettlementID }
func (e *SettlementApprovedEvent) OccurredAt() time.Time { return e.OccurredAt }

type SettlementRejectedEvent struct {
	SettlementID string    `json:"settlement_id"`
	MerchantID   uint64    `json:"merchant_id"`
	Reason       string    `json:"reason"`
	OccurredAt   time.Time `json:"occurred_at"`
}

func (e *SettlementRejectedEvent) EventType() string    { return "settlement.rejected" }
func (e *SettlementRejectedEvent) AggregateID() string  { return e.SettlementID }
func (e *SettlementRejectedEvent) OccurredAt() time.Time { return e.OccurredAt }

type SettlementPaymentStartedEvent struct {
	SettlementID  string          `json:"settlement_id"`
	MerchantID    uint64          `json:"merchant_id"`
	BankAccountID uint64          `json:"bank_account_id"`
	Amount        decimal.Decimal `json:"amount"`
	OccurredAt    time.Time       `json:"occurred_at"`
}

func (e *SettlementPaymentStartedEvent) EventType() string    { return "settlement.payment_started" }
func (e *SettlementPaymentStartedEvent) AggregateID() string  { return e.SettlementID }
func (e *SettlementPaymentStartedEvent) OccurredAt() time.Time { return e.OccurredAt }

type SettlementPaidEvent struct {
	SettlementID   string          `json:"settlement_id"`
	MerchantID     uint64          `json:"merchant_id"`
	Amount         decimal.Decimal `json:"amount"`
	TransactionRef string          `json:"transaction_ref"`
	OccurredAt     time.Time       `json:"occurred_at"`
}

func (e *SettlementPaidEvent) EventType() string    { return "settlement.paid" }
func (e *SettlementPaidEvent) AggregateID() string  { return e.SettlementID }
func (e *SettlementPaidEvent) OccurredAt() time.Time { return e.OccurredAt }

type SettlementPaymentFailedEvent struct {
	SettlementID string    `json:"settlement_id"`
	MerchantID   uint64    `json:"merchant_id"`
	Reason       string    `json:"reason"`
	OccurredAt   time.Time `json:"occurred_at"`
}

func (e *SettlementPaymentFailedEvent) EventType() string    { return "settlement.payment_failed" }
func (e *SettlementPaymentFailedEvent) AggregateID() string  { return e.SettlementID }
func (e *SettlementPaymentFailedEvent) OccurredAt() time.Time { return e.OccurredAt }

type SettlementCancelledEvent struct {
	SettlementID string    `json:"settlement_id"`
	MerchantID   uint64    `json:"merchant_id"`
	Reason       string    `json:"reason"`
	OccurredAt   time.Time `json:"occurred_at"`
}

func (e *SettlementCancelledEvent) EventType() string    { return "settlement.cancelled" }
func (e *SettlementCancelledEvent) AggregateID() string  { return e.SettlementID }
func (e *SettlementCancelledEvent) OccurredAt() time.Time { return e.OccurredAt }

type SettlementAdjustedEvent struct {
	SettlementID     string          `json:"settlement_id"`
	AdjustmentAmount decimal.Decimal `json:"adjustment_amount"`
	Reason           string          `json:"reason"`
	OccurredAt       time.Time       `json:"occurred_at"`
}

func (e *SettlementAdjustedEvent) EventType() string    { return "settlement.adjusted" }
func (e *SettlementAdjustedEvent) AggregateID() string  { return e.SettlementID }
func (e *SettlementAdjustedEvent) OccurredAt() time.Time { return e.OccurredAt }

type BankAccountAddedEvent struct {
	BankAccountID uint64    `json:"bank_account_id"`
	MerchantID    uint64    `json:"merchant_id"`
	BankName      string    `json:"bank_name"`
	AccountNo     string    `json:"account_no"`
	OccurredAt    time.Time `json:"occurred_at"`
}

func (e *BankAccountAddedEvent) EventType() string    { return "bank_account.added" }
func (e *BankAccountAddedEvent) AggregateID() string  { return string(rune(e.BankAccountID)) }
func (e *BankAccountAddedEvent) OccurredAt() time.Time { return e.OccurredAt }

type BankAccountActivatedEvent struct {
	BankAccountID uint64    `json:"bank_account_id"`
	MerchantID    uint64    `json:"merchant_id"`
	OccurredAt    time.Time `json:"occurred_at"`
}

func (e *BankAccountActivatedEvent) EventType() string    { return "bank_account.activated" }
func (e *BankAccountActivatedEvent) AggregateID() string  { return string(rune(e.BankAccountID)) }
func (e *BankAccountActivatedEvent) OccurredAt() time.Time { return e.OccurredAt }
