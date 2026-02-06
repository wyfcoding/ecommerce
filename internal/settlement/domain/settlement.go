package domain

import (
	"context"
	"fmt"
	"time"
)

// --- Settlement Aggregates ---

// SettlementStatus 定义了结算单的生命周期状态。
type SettlementStatus int8

const (
	SettlementStatusPending    SettlementStatus = 0 // 待结算
	SettlementStatusProcessing SettlementStatus = 1 // 结算中
	SettlementStatusCompleted  SettlementStatus = 2 // 已完成
	SettlementStatusFailed     SettlementStatus = 3 // 失败
)

// SettlementCycle 定义了结算的周期类型。
type SettlementCycle string

const (
	SettlementCycleDaily   SettlementCycle = "DAILY"
	SettlementCycleWeekly  SettlementCycle = "WEEKLY"
	SettlementCycleMonthly SettlementCycle = "MONTHLY"
)

// Settlement 实体是结算模块的聚合根。
type Settlement struct {
	ID               uint             `json:"id"`
	CreatedAt        time.Time        `json:"created_at"`
	UpdatedAt        time.Time        `json:"updated_at"`
	SettlementNo     string           `json:"settlement_no"`
	MerchantID       uint64           `json:"merchant_id"`
	Cycle            SettlementCycle  `json:"cycle"`
	StartDate        time.Time        `json:"start_date"`
	EndDate          time.Time        `json:"end_date"`
	OrderCount       int64            `json:"order_count"`
	TotalAmount      uint64           `json:"total_amount"`
	PlatformFee      uint64           `json:"platform_fee"`
	CommissionAmount int64            `json:"commission_amount"`
	RebateAmount     int64            `json:"rebate_amount"`
	OtherFees        int64            `json:"other_fees"`
	SettlementAmount uint64           `json:"settlement_amount"`
	Status           SettlementStatus `json:"status"`
	SettledAt        *time.Time       `json:"settled_at"`
	ApprovedBy       string           `json:"approved_by"`
	ApprovedAt       *time.Time       `json:"approved_at"`
	FailReason       string           `json:"fail_reason"`
}

// SettlementDetail 实体代表结算单中的一个订单明细。
type SettlementDetail struct {
	ID               uint      `json:"id"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
	SettlementID     uint64    `json:"settlement_id"`
	OrderID          uint64    `json:"order_id"`
	OrderNo          string    `json:"order_no"`
	OrderAmount      uint64    `json:"order_amount"`
	PlatformFee      uint64    `json:"platform_fee"`
	LogisticsFee     int64     `json:"logistics_fee"`
	ReturnFee        int64     `json:"return_fee"`
	OtherFee         int64     `json:"other_fee"`
	SettlementAmount uint64    `json:"settlement_amount"`
}

// PaymentStatus 定义了结算支付的生命周期状态。
type PaymentStatus string

const (
	PaymentStatusPending    PaymentStatus = "pending"    // 待处理
	PaymentStatusProcessing PaymentStatus = "processing" // 处理中
	PaymentStatusCompleted  PaymentStatus = "completed"  // 已完成
	PaymentStatusFailed     PaymentStatus = "failed"     // 失败
)

// SettlementPayment 实体代表一笔结算支付记录。
type SettlementPayment struct {
	ID            uint          `json:"id"`
	CreatedAt     time.Time     `json:"created_at"`
	UpdatedAt     time.Time     `json:"updated_at"`
	SettlementID  uint64        `json:"settlement_id"`
	MerchantID    uint64        `json:"merchant_id"`
	Amount        int64         `json:"amount"`
	Status        PaymentStatus `json:"status"`
	TransactionID string        `json:"transaction_id"`
	CompletedAt   *time.Time    `json:"completed_at"`
}

// MerchantAccount 实体代表商户的账户信息。
type MerchantAccount struct {
	ID            uint      `json:"id"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	MerchantID    uint64    `json:"merchant_id"`
	Balance       uint64    `json:"balance"`
	FrozenBalance uint64    `json:"frozen_balance"`
	TotalIncome   uint64    `json:"total_income"`
	TotalWithdraw uint64    `json:"total_withdraw"`
	FeeRate       float64   `json:"fee_rate"`
}

func (a *MerchantAccount) AvailableBalance() uint64 {
	if a.Balance < a.FrozenBalance {
		return 0
	}
	return a.Balance - a.FrozenBalance
}

// --- Ledger Core ---

type AccountType string

const (
	AccountTypeAsset     AccountType = "ASSET"
	AccountTypeLiability AccountType = "LIABILITY"
	AccountTypeEquity    AccountType = "EQUITY"
	AccountTypeIncome    AccountType = "INCOME"
	AccountTypeExpense   AccountType = "EXPENSE"
)

type Direction int

const (
	Debit  Direction = 1
	Credit Direction = -1
)

type Subject struct {
	Code        string      `json:"code"`
	Name        string      `json:"name"`
	Type        AccountType `json:"type"`
	Description string      `json:"description"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
}

type Account struct {
	ID          uint      `json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	SubjectCode string    `json:"subject_code"`
	EntityID    string    `json:"entity_id"`
	Balance     int64     `json:"balance"`
	Currency    string    `json:"currency"`
	Version     int64     `json:"version"`
}

type JournalEntry struct {
	ID            uint        `json:"id"`
	CreatedAt     time.Time   `json:"created_at"`
	UpdatedAt     time.Time   `json:"updated_at"`
	EntryNo       string      `json:"entry_no"`
	TransactionID string      `json:"transaction_id"`
	EventType     string      `json:"event_type"`
	PostingDate   time.Time   `json:"posting_date"`
	Description   string      `json:"description"`
	Lines         []EntryLine `json:"lines"`
}

type EntryLine struct {
	ID          uint      `json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	EntryID     uint64    `json:"entry_id"`
	AccountID   uint64    `json:"account_id"`
	SubjectCode string    `json:"subject_code"`
	Direction   Direction `json:"direction"`
	Amount      int64     `json:"amount"`
}

func (e *JournalEntry) Validate() error {
	var debitSum, creditSum int64
	for _, line := range e.Lines {
		if line.Amount <= 0 {
			return fmt.Errorf("invalid amount in line: %d", line.Amount)
		}
		switch line.Direction {
		case Debit:
			debitSum += line.Amount
		case Credit:
			creditSum += line.Amount
		default:
			return fmt.Errorf("invalid direction")
		}
	}
	if debitSum != creditSum {
		return fmt.Errorf("imbalanced entry: debit=%d, credit=%d", debitSum, creditSum)
	}
	return nil
}

// --- Domain Services ---

// LedgerRepository 账务仓储接口
type LedgerRepository interface {
	// 事务支持
	BeginTx(ctx context.Context) any
	CommitTx(tx any) error
	RollbackTx(tx any) error

	GetSubject(code string) (*Subject, error)
	GetAccount(subjectCode, entityID string) (*Account, error)
	GetAccountByID(id uint64) (*Account, error)
	SaveAccount(account *Account) error
	SaveAccountInTx(ctx context.Context, tx any, account *Account) error
	CreateJournalEntry(entry *JournalEntry) error
	CreateJournalEntryInTx(ctx context.Context, tx any, entry *JournalEntry) error
}

type LedgerService struct {
	repo LedgerRepository
}

// NewLedgerService 创建账务服务。
func NewLedgerService(repo LedgerRepository) *LedgerService {
	return &LedgerService{repo: repo}
}

// PostEntry 过账：校验+保存。
func (s *LedgerService) PostEntry(ctx context.Context, entry *JournalEntry) error {
	if err := entry.Validate(); err != nil {
		return err
	}
	return s.repo.CreateJournalEntry(entry)
}

// CreateAccount 获取或创建账户。
func (s *LedgerService) CreateAccount(ctx context.Context, subjectCode, entityID string) (*Account, error) {
	account, err := s.repo.GetAccount(subjectCode, entityID)
	if err == nil && account != nil {
		return account, nil
	}
	if err != nil {
		return nil, err
	}

	newAcc := &Account{
		SubjectCode: subjectCode,
		EntityID:    entityID,
		Balance:     0,
		Currency:    "CNY",
		Version:     0,
	}
	if err := s.repo.SaveAccount(newAcc); err != nil {
		return nil, err
	}
	return newAcc, nil
}
