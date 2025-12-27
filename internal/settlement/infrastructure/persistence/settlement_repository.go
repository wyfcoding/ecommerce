package persistence

import (
	"context"
	"errors"
	"fmt"

	"github.com/wyfcoding/ecommerce/internal/settlement/domain"
	"gorm.io/gorm"
)

type settlementRepository struct {
	db *gorm.DB
}

// NewSettlementRepository 创建并返回一个新的 settlementRepository 实例。
func NewSettlementRepository(db *gorm.DB) domain.SettlementRepository {
	return &settlementRepository{db: db}
}

// --- 结算单管理 (Settlement methods) ---

func (r *settlementRepository) SaveSettlement(ctx context.Context, settlement *domain.Settlement) error {
	return r.db.WithContext(ctx).Save(settlement).Error
}

func (r *settlementRepository) GetSettlement(ctx context.Context, id uint64) (*domain.Settlement, error) {
	var settlement domain.Settlement
	if err := r.db.WithContext(ctx).First(&settlement, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &settlement, nil
}

func (r *settlementRepository) GetSettlementByNo(ctx context.Context, no string) (*domain.Settlement, error) {
	var settlement domain.Settlement
	if err := r.db.WithContext(ctx).Where("settlement_no = ?", no).First(&settlement).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &settlement, nil
}

func (r *settlementRepository) ListSettlements(ctx context.Context, merchantID uint64, status *domain.SettlementStatus, offset, limit int) ([]*domain.Settlement, int64, error) {
	var list []*domain.Settlement
	var total int64

	db := r.db.WithContext(ctx).Model(&domain.Settlement{})
	if merchantID > 0 {
		db = db.Where("merchant_id = ?", merchantID)
	}
	if status != nil {
		db = db.Where("status = ?", *status)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := db.Offset(offset).Limit(limit).Order("id desc").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// --- 结算明细管理 (SettlementDetail methods) ---

func (r *settlementRepository) SaveSettlementDetail(ctx context.Context, detail *domain.SettlementDetail) error {
	return r.db.WithContext(ctx).Save(detail).Error
}

func (r *settlementRepository) ListSettlementDetails(ctx context.Context, settlementID uint64) ([]*domain.SettlementDetail, error) {
	var list []*domain.SettlementDetail
	if err := r.db.WithContext(ctx).Where("settlement_id = ?", settlementID).Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

// --- 商户账户管理 (MerchantAccount methods) ---

func (r *settlementRepository) GetMerchantAccount(ctx context.Context, merchantID uint64) (*domain.MerchantAccount, error) {
	var account domain.MerchantAccount
	if err := r.db.WithContext(ctx).Where("merchant_id = ?", merchantID).First(&account).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &account, nil
}

func (r *settlementRepository) SaveMerchantAccount(ctx context.Context, account *domain.MerchantAccount) error {
	return r.db.WithContext(ctx).Save(account).Error
}

// --- Ledger Implementation ---

type ledgerRepository struct {
	db *gorm.DB
}

func NewLedgerRepository(db *gorm.DB) domain.LedgerRepository {
	return &ledgerRepository{db: db}
}

func (r *ledgerRepository) GetSubject(code string) (*domain.Subject, error) {
	var s domain.Subject
	if err := r.db.Where("code = ?", code).First(&s).Error; err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *ledgerRepository) GetAccount(subjectCode, entityID string) (*domain.Account, error) {
	var acc domain.Account
	if err := r.db.Where("subject_code = ? AND entity_id = ?", subjectCode, entityID).First(&acc).Error; err != nil {
		return nil, err
	}
	return &acc, nil
}

func (r *ledgerRepository) GetAccountByID(id uint64) (*domain.Account, error) {
	var acc domain.Account
	if err := r.db.First(&acc, id).Error; err != nil {
		return nil, err
	}
	return &acc, nil
}

func (r *ledgerRepository) SaveAccount(account *domain.Account) error {
	return r.db.Save(account).Error
}

func (r *ledgerRepository) CreateJournalEntry(entry *domain.JournalEntry) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(entry).Error; err != nil {
			return err
		}

		for _, line := range entry.Lines {
			var acc domain.Account
			if err := tx.Set("gorm:query_option", "FOR UPDATE").First(&acc, line.AccountID).Error; err != nil {
				return err
			}

			var subject domain.Subject
			if err := tx.Where("code = ?", acc.SubjectCode).First(&subject).Error; err != nil {
				return err
			}

			changeAmount := line.Amount
			isDebit := line.Direction == domain.Debit

			switch subject.Type {
			case domain.AccountTypeAsset, domain.AccountTypeExpense:
				if !isDebit {
					changeAmount = -changeAmount
				}
			case domain.AccountTypeLiability, domain.AccountTypeEquity, domain.AccountTypeIncome:
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
