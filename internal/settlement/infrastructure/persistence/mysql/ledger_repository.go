package mysql

import (
	"context"

	"github.com/wyfcoding/ecommerce/internal/settlement/domain"
	"gorm.io/gorm"
)

type ledgerRepository struct {
	db *gorm.DB
}

// NewLedgerRepository 创建账务仓储实现。
func NewLedgerRepository(db *gorm.DB) domain.LedgerRepository {
	return &ledgerRepository{db: db}
}

func (r *ledgerRepository) BeginTx(ctx context.Context) any {
	return r.db.WithContext(ctx).Begin()
}

func (r *ledgerRepository) CommitTx(tx any) error {
	return tx.(*gorm.DB).Commit().Error
}

func (r *ledgerRepository) RollbackTx(tx any) error {
	return tx.(*gorm.DB).Rollback().Error
}

func (r *ledgerRepository) GetSubject(code string) (*domain.Subject, error) {
	var model SubjectModel
	if err := r.db.Where("code = ?", code).First(&model).Error; err != nil {
		return nil, err
	}
	return toSubject(&model), nil
}

func (r *ledgerRepository) GetAccount(subjectCode, entityID string) (*domain.Account, error) {
	var model AccountModel
	if err := r.db.Where("subject_code = ? AND entity_id = ?", subjectCode, entityID).First(&model).Error; err != nil {
		return nil, err
	}
	return toAccount(&model), nil
}

func (r *ledgerRepository) GetAccountByID(id uint64) (*domain.Account, error) {
	var model AccountModel
	if err := r.db.First(&model, id).Error; err != nil {
		return nil, err
	}
	return toAccount(&model), nil
}

func (r *ledgerRepository) SaveAccount(account *domain.Account) error {
	model := toAccountModel(account)
	if model == nil {
		return nil
	}
	if err := r.db.Save(model).Error; err != nil {
		return err
	}
	*account = *toAccount(model)
	return nil
}

func (r *ledgerRepository) SaveAccountInTx(ctx context.Context, tx any, account *domain.Account) error {
	model := toAccountModel(account)
	if model == nil {
		return nil
	}
	if err := tx.(*gorm.DB).WithContext(ctx).Save(model).Error; err != nil {
		return err
	}
	*account = *toAccount(model)
	return nil
}

func (r *ledgerRepository) CreateJournalEntry(entry *domain.JournalEntry) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		return r.CreateJournalEntryInTx(context.Background(), tx, entry)
	})
}

func (r *ledgerRepository) CreateJournalEntryInTx(ctx context.Context, tx any, entry *domain.JournalEntry) error {
	model := toJournalEntryModel(entry)
	if model == nil {
		return nil
	}
	db := tx.(*gorm.DB).WithContext(ctx)
	if err := db.Create(model).Error; err != nil {
		return err
	}
	*entry = *toJournalEntry(model)
	return nil
}

func (r *ledgerRepository) Save(ctx context.Context, ledger *domain.Ledger) error {
	model := toLedgerModel(ledger)
	if model == nil {
		return nil
	}
	if err := r.db.WithContext(ctx).Save(model).Error; err != nil {
		return err
	}
	*ledger = *toLedger(model)
	return nil
}

func (r *ledgerRepository) GetByID(ctx context.Context, id uint64) (*domain.Ledger, error) {
	var model LedgerModel
	if err := r.db.WithContext(ctx).First(&model, id).Error; err != nil {
		return nil, err
	}
	return toLedger(&model), nil
}

func (r *ledgerRepository) GetBySettlementID(ctx context.Context, settlementID uint64) (*domain.Ledger, error) {
	var model LedgerModel
	if err := r.db.WithContext(ctx).Where("settlement_id = ?", settlementID).First(&model).Error; err != nil {
		return nil, err
	}
	return toLedger(&model), nil
}

func (r *ledgerRepository) ListByMerchant(ctx context.Context, merchantID uint64, offset, limit int) ([]*domain.Ledger, int64, error) {
	var models []LedgerModel
	var count int64
	db := r.db.WithContext(ctx).Model(&LedgerModel{}).Where("merchant_id = ?", merchantID)
	if err := db.Count(&count).Error; err != nil {
		return nil, 0, err
	}
	if err := db.Offset(offset).Limit(limit).Find(&models).Error; err != nil {
		return nil, 0, err
	}
	ledgers := make([]*domain.Ledger, 0, len(models))
	for i := range models {
		ledgers = append(ledgers, toLedger(&models[i]))
	}
	return ledgers, count, nil
}
