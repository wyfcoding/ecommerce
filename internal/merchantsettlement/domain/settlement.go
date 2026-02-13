package domain

import (
	"time"

	"github.com/shopspring/decimal"
)

type SettlementStatus string

const (
	StatusPending    SettlementStatus = "PENDING"
	StatusCalculating SettlementStatus = "CALCULATING"
	StatusPendingApproval SettlementStatus = "PENDING_APPROVAL"
	StatusApproved   SettlementStatus = "APPROVED"
	StatusPaying     SettlementStatus = "PAYING"
	StatusPaid       SettlementStatus = "PAID"
	StatusFailed     SettlementStatus = "FAILED"
	StatusCancelled  SettlementStatus = "CANCELLED"
)

type SettlementCycle string

const (
	CycleDaily   SettlementCycle = "DAILY"
	CycleWeekly  SettlementCycle = "WEEKLY"
	CycleMonthly SettlementCycle = "MONTHLY"
)

type Settlement struct {
	ID               uint64           `json:"id"`
	SettlementID     string           `json:"settlement_id"`
	MerchantID       uint64           `json:"merchant_id"`
	Cycle            SettlementCycle  `json:"cycle"`
	PeriodStart      time.Time        `json:"period_start"`
	PeriodEnd        time.Time        `json:"period_end"`
	OrderCount       int64            `json:"order_count"`
	GrossAmount      decimal.Decimal  `json:"gross_amount"`
	RefundAmount     decimal.Decimal  `json:"refund_amount"`
	PlatformCommission decimal.Decimal `json:"platform_commission"`
	PromotionFee     decimal.Decimal  `json:"promotion_fee"`
	LogisticsFee     decimal.Decimal  `json:"logistics_fee"`
	AdjustmentAmount decimal.Decimal  `json:"adjustment_amount"`
	SettlementAmount decimal.Decimal  `json:"settlement_amount"`
	Status           SettlementStatus `json:"status"`
	BankAccountID    uint64           `json:"bank_account_id"`
	TransactionRef   string           `json:"transaction_ref"`
	ApprovedBy       uint64           `json:"approved_by"`
	ApprovedAt       *time.Time       `json:"approved_at"`
	PaidAt           *time.Time       `json:"paid_at"`
	FailReason       string           `json:"fail_reason"`
	Version          int64            `json:"version"`
	CreatedAt        time.Time        `json:"created_at"`
	UpdatedAt        time.Time        `json:"updated_at"`
	details          []*SettlementDetail
	events           []DomainEvent
}

func NewSettlement(settlementID string, merchantID uint64, cycle SettlementCycle, periodStart, periodEnd time.Time) *Settlement {
	return &Settlement{
		SettlementID:     settlementID,
		MerchantID:       merchantID,
		Cycle:            cycle,
		PeriodStart:      periodStart,
		PeriodEnd:        periodEnd,
		GrossAmount:      decimal.Zero,
		RefundAmount:     decimal.Zero,
		PlatformCommission: decimal.Zero,
		PromotionFee:     decimal.Zero,
		LogisticsFee:     decimal.Zero,
		AdjustmentAmount: decimal.Zero,
		SettlementAmount: decimal.Zero,
		Status:           StatusPending,
		Version:          0,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
		details:          make([]*SettlementDetail, 0),
		events:           make([]DomainEvent, 0),
	}
}

func (s *Settlement) AddDetail(detail *SettlementDetail) {
	s.details = append(s.details, detail)
	s.OrderCount++
	s.GrossAmount = s.GrossAmount.Add(detail.OrderAmount)
	s.RefundAmount = s.RefundAmount.Add(detail.RefundAmount)
	s.PlatformCommission = s.PlatformCommission.Add(detail.PlatformCommission)
	s.PromotionFee = s.PromotionFee.Add(detail.PromotionFee)
	s.LogisticsFee = s.LogisticsFee.Add(detail.LogisticsFee)
	s.recalculate()
}

func (s *Settlement) recalculate() {
	s.SettlementAmount = s.GrossAmount.
		Sub(s.RefundAmount).
		Sub(s.PlatformCommission).
		Sub(s.PromotionFee).
		Sub(s.LogisticsFee).
		Add(s.AdjustmentAmount)
	s.UpdatedAt = time.Now()
}

func (s *Settlement) SetAdjustment(amount decimal.Decimal, reason string) {
	s.AdjustmentAmount = amount
	s.recalculate()
	s.addEvent(&SettlementAdjustedEvent{
		SettlementID:     s.SettlementID,
		AdjustmentAmount: amount,
		Reason:           reason,
		OccurredAt:       time.Now(),
	})
}

func (s *Settlement) StartCalculation() error {
	if s.Status != StatusPending {
		return ErrInvalidStatus
	}
	s.Status = StatusCalculating
	s.UpdatedAt = time.Now()
	s.addEvent(&SettlementCalculationStartedEvent{
		SettlementID: s.SettlementID,
		MerchantID:   s.MerchantID,
		OccurredAt:   time.Now(),
	})
	return nil
}

func (s *Settlement) CompleteCalculation() error {
	if s.Status != StatusCalculating {
		return ErrInvalidStatus
	}
	s.Status = StatusPendingApproval
	s.UpdatedAt = time.Now()
	s.addEvent(&SettlementCalculationCompletedEvent{
		SettlementID:     s.SettlementID,
		MerchantID:       s.MerchantID,
		SettlementAmount: s.SettlementAmount,
		OccurredAt:       time.Now(),
	})
	return nil
}

func (s *Settlement) Approve(approvedBy uint64) error {
	if s.Status != StatusPendingApproval {
		return ErrInvalidStatus
	}
	now := time.Now()
	s.Status = StatusApproved
	s.ApprovedBy = approvedBy
	s.ApprovedAt = &now
	s.UpdatedAt = now
	s.addEvent(&SettlementApprovedEvent{
		SettlementID: s.SettlementID,
		MerchantID:   s.MerchantID,
		ApprovedBy:   approvedBy,
		OccurredAt:   now,
	})
	return nil
}

func (s *Settlement) Reject(reason string) error {
	if s.Status != StatusPendingApproval {
		return ErrInvalidStatus
	}
	s.Status = StatusPending
	s.FailReason = reason
	s.UpdatedAt = time.Now()
	s.addEvent(&SettlementRejectedEvent{
		SettlementID: s.SettlementID,
		MerchantID:   s.MerchantID,
		Reason:       reason,
		OccurredAt:   time.Now(),
	})
	return nil
}

func (s *Settlement) StartPayment(bankAccountID uint64) error {
	if s.Status != StatusApproved {
		return ErrInvalidStatus
	}
	s.Status = StatusPaying
	s.BankAccountID = bankAccountID
	s.UpdatedAt = time.Now()
	s.addEvent(&SettlementPaymentStartedEvent{
		SettlementID:   s.SettlementID,
		MerchantID:     s.MerchantID,
		BankAccountID:  bankAccountID,
		Amount:         s.SettlementAmount,
		OccurredAt:     time.Now(),
	})
	return nil
}

func (s *Settlement) CompletePayment(transactionRef string) error {
	if s.Status != StatusPaying {
		return ErrInvalidStatus
	}
	now := time.Now()
	s.Status = StatusPaid
	s.TransactionRef = transactionRef
	s.PaidAt = &now
	s.UpdatedAt = now
	s.addEvent(&SettlementPaidEvent{
		SettlementID:   s.SettlementID,
		MerchantID:     s.MerchantID,
		Amount:         s.SettlementAmount,
		TransactionRef: transactionRef,
		OccurredAt:     now,
	})
	return nil
}

func (s *Settlement) FailPayment(reason string) error {
	if s.Status != StatusPaying {
		return ErrInvalidStatus
	}
	s.Status = StatusFailed
	s.FailReason = reason
	s.UpdatedAt = time.Now()
	s.addEvent(&SettlementPaymentFailedEvent{
		SettlementID: s.SettlementID,
		MerchantID:   s.MerchantID,
		Reason:       reason,
		OccurredAt:   time.Now(),
	})
	return nil
}

func (s *Settlement) Cancel(reason string) error {
	if s.Status == StatusPaid || s.Status == StatusPaying {
		return ErrInvalidStatus
	}
	s.Status = StatusCancelled
	s.FailReason = reason
	s.UpdatedAt = time.Now()
	s.addEvent(&SettlementCancelledEvent{
		SettlementID: s.SettlementID,
		MerchantID:   s.MerchantID,
		Reason:       reason,
		OccurredAt:   time.Now(),
	})
	return nil
}

func (s *Settlement) Details() []*SettlementDetail {
	return s.details
}

func (s *Settlement) Events() []DomainEvent {
	return s.events
}

func (s *Settlement) ClearEvents() {
	s.events = make([]DomainEvent, 0)
}

func (s *Settlement) addEvent(event DomainEvent) {
	s.events = append(s.events, event)
}

type SettlementDetail struct {
	ID                 uint64          `json:"id"`
	SettlementID       string          `json:"settlement_id"`
	OrderID            uint64          `json:"order_id"`
	OrderNo            string          `json:"order_no"`
	OrderAmount        decimal.Decimal `json:"order_amount"`
	RefundAmount       decimal.Decimal `json:"refund_amount"`
	PlatformCommission decimal.Decimal `json:"platform_commission"`
	PromotionFee       decimal.Decimal `json:"promotion_fee"`
	LogisticsFee       decimal.Decimal `json:"logistics_fee"`
	SettlementAmount   decimal.Decimal `json:"settlement_amount"`
	CreatedAt          time.Time       `json:"created_at"`
}

func NewSettlementDetail(settlementID string, orderID uint64, orderNo string, orderAmount, refundAmount, platformCommission, promotionFee, logisticsFee decimal.Decimal) *SettlementDetail {
	settlementAmount := orderAmount.Sub(refundAmount).Sub(platformCommission).Sub(promotionFee).Sub(logisticsFee)
	return &SettlementDetail{
		SettlementID:       settlementID,
		OrderID:            orderID,
		OrderNo:            orderNo,
		OrderAmount:        orderAmount,
		RefundAmount:       refundAmount,
		PlatformCommission: platformCommission,
		PromotionFee:       promotionFee,
		LogisticsFee:       logisticsFee,
		SettlementAmount:   settlementAmount,
		CreatedAt:          time.Now(),
	}
}

type MerchantBankAccount struct {
	ID           uint64          `json:"id"`
	MerchantID   uint64          `json:"merchant_id"`
	BankName     string          `json:"bank_name"`
	BankCode     string          `json:"bank_code"`
	AccountName  string          `json:"account_name"`
	AccountNo    string          `json:"account_no"`
	BranchName   string          `json:"branch_name"`
	IsDefault    bool            `json:"is_default"`
	Status       AccountStatus   `json:"status"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
}

type AccountStatus string

const (
	AccountStatusActive   AccountStatus = "ACTIVE"
	AccountStatusInactive AccountStatus = "INACTIVE"
)

type MerchantSettlementConfig struct {
	ID                  uint64          `json:"id"`
	MerchantID          uint64          `json:"merchant_id"`
	Cycle               SettlementCycle `json:"cycle"`
	CommissionRate      decimal.Decimal `json:"commission_rate"`
	MinSettlementAmount decimal.Decimal `json:"min_settlement_amount"`
	AutoApprove         bool            `json:"auto_approve"`
	AutoPay             bool            `json:"auto_pay"`
	Status              ConfigStatus    `json:"status"`
	CreatedAt           time.Time       `json:"created_at"`
	UpdatedAt           time.Time       `json:"updated_at"`
}

type ConfigStatus string

const (
	ConfigStatusActive   ConfigStatus = "ACTIVE"
	ConfigStatusInactive ConfigStatus = "INACTIVE"
)
