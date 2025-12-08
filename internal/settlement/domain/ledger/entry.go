package ledger

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

// JournalEntry 日记账分录 (凭证)
// 代表一次原子性的财务变动，必须满足 有借必有贷，借贷必相等
type JournalEntry struct {
	gorm.Model
	EntryNo       string    `gorm:"uniqueIndex;type:varchar(64);not null;comment:凭证号" json:"entry_no"`
	TransactionID string    `gorm:"index;type:varchar(64);not null;comment:关联业务流水号" json:"transaction_id"` // 关联的支付ID或退款ID
	EventType     string    `gorm:"type:varchar(32);not null;comment:事件类型" json:"event_type"`              // e.g., "PAYMENT_SUCCESS", "REFUND"
	PostingDate   time.Time `gorm:"index;not null;comment:入账日期" json:"posting_date"`
	Description   string    `gorm:"type:varchar(255);comment:摘要" json:"description"`

	Lines []EntryLine `gorm:"foreignKey:EntryID" json:"lines"` // 分录明细
}

// EntryLine 分录明细
type EntryLine struct {
	gorm.Model
	EntryID     uint64    `gorm:"index;not null;comment:关联凭证ID" json:"entry_id"`
	AccountID   uint64    `gorm:"index;not null;comment:关联账户ID" json:"account_id"`
	SubjectCode string    `gorm:"type:varchar(32);not null;comment:冗余科目代码" json:"subject_code"`
	Direction   Direction `gorm:"type:tinyint;not null;comment:借贷方向(1:借, -1:贷)" json:"direction"`
	Amount      int64     `gorm:"not null;comment:发生金额(分)" json:"amount"`
}

// Validate 校验凭证平衡性
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
