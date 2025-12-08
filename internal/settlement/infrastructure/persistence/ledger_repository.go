package persistence

import (
	"fmt"

	"github.com/wyfcoding/ecommerce/internal/settlement/domain/ledger"
	"gorm.io/gorm"
)

type LedgerRepositoryImpl struct {
	db *gorm.DB
}

func NewLedgerRepository(db *gorm.DB) ledger.LedgerRepository {
	return &LedgerRepositoryImpl{db: db}
}

func (r *LedgerRepositoryImpl) GetSubject(code string) (*ledger.Subject, error) {
	var s ledger.Subject
	if err := r.db.Where("code = ?", code).First(&s).Error; err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *LedgerRepositoryImpl) GetAccount(subjectCode, entityID string) (*ledger.Account, error) {
	var acc ledger.Account
	if err := r.db.Where("subject_code = ? AND entity_id = ?", subjectCode, entityID).First(&acc).Error; err != nil {
		return nil, err
	}
	return &acc, nil
}

func (r *LedgerRepositoryImpl) GetAccountByID(id uint64) (*ledger.Account, error) {
	var acc ledger.Account
	if err := r.db.First(&acc, id).Error; err != nil {
		return nil, err
	}
	return &acc, nil
}

func (r *LedgerRepositoryImpl) SaveAccount(account *ledger.Account) error {
	return r.db.Save(account).Error
}

// CreateJournalEntry 保存凭证并原子更新余额
func (r *LedgerRepositoryImpl) CreateJournalEntry(entry *ledger.JournalEntry) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 1. 保存凭证主表
		if err := tx.Create(entry).Error; err != nil {
			return err
		}

		// 2. 逐行更新账户余额
		// 借(Debit) = 余额增加 (Asset/Expense) 或 减少 (Liability/Equity/Income)
		// 贷(Credit) = 相反
		// 这里我们简化模型：我们在 Subject 定义中应包含借贷方向属性（通常称借方余额科目/贷方余额科目）
		// 为了简化，假设 Account Balance 代表 "该科目方向的净值"
		// Asset: Debit +, Credit -
		// Liability: Credit +, Debit -

		for _, line := range entry.Lines {
			var acc ledger.Account
			if err := tx.Set("gorm:query_option", "FOR UPDATE").First(&acc, line.AccountID).Error; err != nil {
				return err
			}

			// 获取科目类型以决定正负
			var subject ledger.Subject
			if err := tx.Where("code = ?", acc.SubjectCode).First(&subject).Error; err != nil {
				return err
			}

			changeAmount := line.Amount

			// 标准会计恒等式逻辑
			isDebit := line.Direction == ledger.Debit

			switch subject.Type {
			case ledger.AccountTypeAsset, ledger.AccountTypeExpense:
				// 借增贷减
				if !isDebit {
					changeAmount = -changeAmount
				}
			case ledger.AccountTypeLiability, ledger.AccountTypeEquity, ledger.AccountTypeIncome:
				// 贷增借减
				if isDebit {
					changeAmount = -changeAmount
				}
			default:
				return fmt.Errorf("unknown account type: %s", subject.Type)
			}

			acc.Balance += changeAmount

			if err := tx.Save(&acc).Error; err != nil {
				return err
			}
		}

		return nil
	})
}
