package domain

import (
	"gorm.io/gorm"
)

// AccountType 账户/科目类型
type AccountType string

const (
	AccountTypeAsset     AccountType = "ASSET"     // 资产类 (e.g., 银行存款, 应收账款)
	AccountTypeLiability AccountType = "LIABILITY" // 负债类 (e.g., 应付商户款, 用户余额)
	AccountTypeEquity    AccountType = "EQUITY"    // 权益类 (e.g., 留存收益)
	AccountTypeIncome    AccountType = "INCOME"    // 收入类 (e.g., 手续费收入)
	AccountTypeExpense   AccountType = "EXPENSE"   // 费用类 (e.g., 通道成本)
)

// Direction 借贷方向
type Direction int

const (
	Debit  Direction = 1  // 借
	Credit Direction = -1 // 贷
)

// Subject 会计科目
// 定义了资金的分类，如 "2001: 应付账款-商户", "1002: 银行存款-支付宝"
type Subject struct {
	Code        string      `gorm:"primaryKey;type:varchar(32);comment:科目代码" json:"code"`
	Name        string      `gorm:"type:varchar(64);not null;comment:科目名称" json:"name"`
	Type        AccountType `gorm:"type:varchar(32);not null;comment:科目类型" json:"type"`
	Description string      `gorm:"type:varchar(255);comment:描述" json:"description"`
}

// Account 财务账户
// 具体的资金载体，通常是 科目 + 实体ID 的组合
// 例如：科目="2001(应付商户)", EntityID="Merch_1001" -> 商户1001的待结算户
type Account struct {
	gorm.Model
	SubjectCode string `gorm:"index;type:varchar(32);not null;comment:关联科目" json:"subject_code"`
	EntityID    string `gorm:"uniqueIndex:idx_sub_ent;type:varchar(64);not null;comment:关联实体ID" json:"entity_id"` // e.g., "USER_1", "MERCH_1001"
	Balance     int64  `gorm:"not null;default:0;comment:余额(分)" json:"balance"`                                   // 当前余额（分）
	Currency    string `gorm:"type:varchar(3);default:'CNY';comment:币种" json:"currency"`
	Version     int64  `gorm:"default:0;comment:乐观锁版本" json:"version"` // 乐观锁
}

// LedgerRepository 账务仓储接口
type LedgerRepository interface {
	GetSubject(code string) (*Subject, error)
	GetAccount(subjectCode, entityID string) (*Account, error)
	GetAccountByID(id uint64) (*Account, error)
	SaveAccount(account *Account) error
	CreateJournalEntry(entry *JournalEntry) error
}
