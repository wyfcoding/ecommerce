package domain

import (
	"time"

	"github.com/shopspring/decimal"
)

type DomainEvent interface {
	EventType() string
	OccurredAtTime() time.Time
	AggregateID() string
}

type TransactionCreatedEvent struct {
	TransactionID string          `json:"transaction_id"`
	UserID        uint64          `json:"user_id"`
	Type          TransactionType `json:"type"`
	Amount        decimal.Decimal `json:"amount"`
	Currency      string          `json:"currency"`
	Channel       ChannelType     `json:"channel"`
	OccurredAt    time.Time       `json:"occurred_at"`
}

func (e *TransactionCreatedEvent) EventType() string   { return "fiat.transaction.created" }
func (e *TransactionCreatedEvent) AggregateID() string { return e.TransactionID }
func (e *TransactionCreatedEvent) OccurredAtTime() time.Time { return e.OccurredAt }

type TransactionProcessingEvent struct {
	TransactionID string    `json:"transaction_id"`
	UserID        uint64    `json:"user_id"`
	OccurredAt    time.Time `json:"occurred_at"`
}

func (e *TransactionProcessingEvent) EventType() string   { return "fiat.transaction.processing" }
func (e *TransactionProcessingEvent) AggregateID() string { return e.TransactionID }
func (e *TransactionProcessingEvent) OccurredAtTime() time.Time { return e.OccurredAt }

type TransactionCompletedEvent struct {
	TransactionID string          `json:"transaction_id"`
	UserID        uint64          `json:"user_id"`
	Amount        decimal.Decimal `json:"amount"`
	Currency      string          `json:"currency"`
	ExternalTxID  string          `json:"external_tx_id"`
	OccurredAt    time.Time       `json:"occurred_at"`
}

func (e *TransactionCompletedEvent) EventType() string   { return "fiat.transaction.completed" }
func (e *TransactionCompletedEvent) AggregateID() string { return e.TransactionID }
func (e *TransactionCompletedEvent) OccurredAtTime() time.Time { return e.OccurredAt }

type TransactionFailedEvent struct {
	TransactionID string    `json:"transaction_id"`
	UserID        uint64    `json:"user_id"`
	Reason        string    `json:"reason"`
	OccurredAt    time.Time `json:"occurred_at"`
}

func (e *TransactionFailedEvent) EventType() string   { return "fiat.transaction.failed" }
func (e *TransactionFailedEvent) AggregateID() string { return e.TransactionID }
func (e *TransactionFailedEvent) OccurredAtTime() time.Time { return e.OccurredAt }

type TransactionCancelledEvent struct {
	TransactionID string    `json:"transaction_id"`
	UserID        uint64    `json:"user_id"`
	Reason        string    `json:"reason"`
	OccurredAt    time.Time `json:"occurred_at"`
}

func (e *TransactionCancelledEvent) EventType() string   { return "fiat.transaction.cancelled" }
func (e *TransactionCancelledEvent) AggregateID() string { return e.TransactionID }
func (e *TransactionCancelledEvent) OccurredAtTime() time.Time { return e.OccurredAt }

type TransactionRefundedEvent struct {
	TransactionID string          `json:"transaction_id"`
	UserID        uint64          `json:"user_id"`
	Amount        decimal.Decimal `json:"amount"`
	Currency      string          `json:"currency"`
	OccurredAt    time.Time       `json:"occurred_at"`
}

func (e *TransactionRefundedEvent) EventType() string   { return "fiat.transaction.refunded" }
func (e *TransactionRefundedEvent) AggregateID() string { return e.TransactionID }
func (e *TransactionRefundedEvent) OccurredAtTime() time.Time { return e.OccurredAt }

type ExchangeRateUpdatedEvent struct {
	FromCurrency string          `json:"from_currency"`
	ToCurrency   string          `json:"to_currency"`
	OldRate      decimal.Decimal `json:"old_rate"`
	NewRate      decimal.Decimal `json:"new_rate"`
	Source       string          `json:"source"`
	OccurredAt   time.Time       `json:"occurred_at"`
}

func (e *ExchangeRateUpdatedEvent) EventType() string { return "fiat.exchange_rate.updated" }
func (e *ExchangeRateUpdatedEvent) AggregateID() string { return e.FromCurrency + ":" + e.ToCurrency }
func (e *ExchangeRateUpdatedEvent) OccurredAtTime() time.Time { return e.OccurredAt }

type BankAccountAddedEvent struct {
	BankAccountID uint64    `json:"bank_account_id"`
	UserID        uint64    `json:"user_id"`
	BankName      string    `json:"bank_name"`
	Currency      string    `json:"currency"`
	OccurredAt    time.Time `json:"occurred_at"`
}

func (e *BankAccountAddedEvent) EventType() string   { return "fiat.bank_account.added" }
func (e *BankAccountAddedEvent) AggregateID() string { return string(rune(e.BankAccountID)) }
func (e *BankAccountAddedEvent) OccurredAtTime() time.Time { return e.OccurredAt }

type BankAccountVerifiedEvent struct {
	BankAccountID uint64    `json:"bank_account_id"`
	UserID        uint64    `json:"user_id"`
	OccurredAt    time.Time `json:"occurred_at"`
}

func (e *BankAccountVerifiedEvent) EventType() string   { return "fiat.bank_account.verified" }
func (e *BankAccountVerifiedEvent) AggregateID() string { return string(rune(e.BankAccountID)) }
func (e *BankAccountVerifiedEvent) OccurredAtTime() time.Time { return e.OccurredAt }
