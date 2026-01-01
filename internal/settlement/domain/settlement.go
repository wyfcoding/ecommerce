package domain

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
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
	gorm.Model
	SettlementNo     string           `gorm:"type:varchar(64);uniqueIndex;not null;comment:结算单号" json:"settlement_no"`
	MerchantID       uint64           `gorm:"index;not null;comment:商户ID" json:"merchant_id"`
	Cycle            SettlementCycle  `gorm:"type:varchar(32);not null;comment:结算周期" json:"cycle"`
	StartDate        time.Time        `gorm:"not null;comment:开始日期" json:"start_date"`
	EndDate          time.Time        `gorm:"not null;comment:结束日期" json:"end_date"`
	OrderCount       int64            `gorm:"not null;default:0;comment:订单数量" json:"order_count"`
	TotalAmount      uint64           `gorm:"not null;default:0;comment:总金额(分)" json:"total_amount"`
	PlatformFee      uint64           `gorm:"not null;default:0;comment:平台手续费(分)" json:"platform_fee"`
	SettlementAmount uint64           `gorm:"not null;default:0;comment:结算金额(分)" json:"settlement_amount"`
	Status           SettlementStatus `gorm:"type:tinyint;not null;default:0;comment:状态" json:"status"`
	SettledAt        *time.Time       `gorm:"comment:结算时间" json:"settled_at"`
	FailReason       string           `gorm:"type:varchar(255);comment:失败原因" json:"fail_reason"`
}

// SettlementDetail 实体代表结算单中的一个订单明细。
type SettlementDetail struct {
	gorm.Model
	SettlementID     uint64 `gorm:"index;not null;comment:结算单ID" json:"settlement_id"`
	OrderID          uint64 `gorm:"index;not null;comment:订单ID" json:"order_id"`
	OrderNo          string `gorm:"type:varchar(64);not null;comment:订单号" json:"order_no"`
	OrderAmount      uint64 `gorm:"not null;comment:订单金额(分)" json:"order_amount"`
	PlatformFee      uint64 `gorm:"not null;comment:平台手续费(分)" json:"platform_fee"`
	SettlementAmount uint64 `gorm:"not null;comment:结算金额(分)" json:"settlement_amount"`
}

// MerchantAccount 实体代表商户的账户信息。
type MerchantAccount struct {
	gorm.Model
	MerchantID    uint64  `gorm:"uniqueIndex;not null;comment:商户ID" json:"merchant_id"`
	Balance       uint64  `gorm:"not null;default:0;comment:余额(分)" json:"balance"`
	FrozenBalance uint64  `gorm:"not null;default:0;comment:冻结金额(分)" json:"frozen_balance"`
	TotalIncome   uint64  `gorm:"not null;default:0;comment:总收入(分)" json:"total_income"`
	TotalWithdraw uint64  `gorm:"not null;default:0;comment:总提现(分)" json:"total_withdraw"`
	FeeRate       float64 `gorm:"type:decimal(5,2);not null;default:0;comment:费率(%)" json:"fee_rate"`
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
	Code        string      `gorm:"primaryKey;type:varchar(32);comment:科目代码" json:"code"`
	Name        string      `gorm:"type:varchar(64);not null;comment:科目名称" json:"name"`
	Type        AccountType `gorm:"type:varchar(32);not null;comment:科目类型" json:"type"`
	Description string      `gorm:"type:varchar(255);comment:描述" json:"description"`
}

type Account struct {
	gorm.Model
	SubjectCode string `gorm:"index;type:varchar(32);not null;comment:关联科目" json:"subject_code"`
	EntityID    string `gorm:"uniqueIndex:idx_sub_ent;type:varchar(64);not null;comment:关联实体ID" json:"entity_id"`
	Balance     int64  `gorm:"not null;default:0;comment:余额(分)" json:"balance"`
	Currency    string `gorm:"type:varchar(3);default:'CNY';comment:币种" json:"currency"`
	Version     int64  `gorm:"default:0;comment:乐观锁版本" json:"version"`
}

type JournalEntry struct {
	gorm.Model
	EntryNo       string      `gorm:"uniqueIndex;type:varchar(64);not null;comment:凭证号" json:"entry_no"`
	TransactionID string      `gorm:"index;type:varchar(64);not null;comment:关联业务流水号" json:"transaction_id"`
	EventType     string      `gorm:"type:varchar(32);not null;comment:事件类型" json:"event_type"`
	PostingDate   time.Time   `gorm:"index;not null;comment:入账日期" json:"posting_date"`
	Description   string      `gorm:"type:varchar(255);comment:摘要" json:"description"`
	Lines         []EntryLine `gorm:"foreignKey:EntryID" json:"lines"`
}

type EntryLine struct {
	gorm.Model
	EntryID     uint64    `gorm:"index;not null;comment:关联凭证ID" json:"entry_id"`
	AccountID   uint64    `gorm:"index;not null;comment:关联账户ID" json:"account_id"`
	SubjectCode string    `gorm:"type:varchar(32);not null;comment:冗余科目代码" json:"subject_code"`
	Direction   Direction `gorm:"type:tinyint;not null;comment:借贷方向(1:借, -1:贷)" json:"direction"`
	Amount      int64     `gorm:"not null;comment:发生金额(分)" json:"amount"`
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
	GetSubject(code string) (*Subject, error)
	GetAccount(subjectCode, entityID string) (*Account, error)
	GetAccountByID(id uint64) (*Account, error)
	SaveAccount(account *Account) error
	CreateJournalEntry(entry *JournalEntry) error
}

type LedgerService struct {
	repo LedgerRepository
}

func NewLedgerService(repo LedgerRepository) *LedgerService {
	return &LedgerService{repo: repo}
}

func (s *LedgerService) PostEntry(ctx context.Context, entry *JournalEntry) error {
	if err := entry.Validate(); err != nil {
		return err
	}
	if entry.EntryNo == "" {
		entry.EntryNo = fmt.Sprintf("JE%d", time.Now().UnixNano())
	}
	if entry.PostingDate.IsZero() {
		entry.PostingDate = time.Now()
	}
	return s.repo.CreateJournalEntry(entry)
}

func (s *LedgerService) CreateAccount(ctx context.Context, subjectCode, entityID string) (*Account, error) {
	acc, err := s.repo.GetAccount(subjectCode, entityID)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing account: %w", err)
	}
	if acc != nil {
		return acc, nil
	}
	acc = &Account{
		SubjectCode: subjectCode,
		EntityID:    entityID,
		Balance:     0,
		Currency:    "CNY",
	}
	if err = s.repo.SaveAccount(acc); err != nil {
		return nil, err
	}
	return acc, nil
}
