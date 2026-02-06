package mysql

import (
	"context"
	"errors"

	"github.com/wyfcoding/ecommerce/internal/loyalty/domain"
	"gorm.io/gorm"
)

type loyaltyRepository struct {
	db *gorm.DB
}

// NewLoyaltyRepository 创建并返回一个新的 loyaltyRepository 实例。
func NewLoyaltyRepository(db *gorm.DB) domain.LoyaltyRepository {
	return &loyaltyRepository{db: db}
}

func (r *loyaltyRepository) BeginTx(ctx context.Context) any {
	return r.db.WithContext(ctx).Begin()
}

func (r *loyaltyRepository) CommitTx(tx any) error {
	gormTx, ok := tx.(*gorm.DB)
	if !ok || gormTx == nil {
		return errors.New("invalid transaction")
	}
	return gormTx.Commit().Error
}

func (r *loyaltyRepository) RollbackTx(tx any) error {
	gormTx, ok := tx.(*gorm.DB)
	if !ok || gormTx == nil {
		return errors.New("invalid transaction")
	}
	return gormTx.Rollback().Error
}

func (r *loyaltyRepository) WithTx(ctx context.Context, fn func(tx any) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(tx)
	})
}

// --- 会员账户 ---

func (r *loyaltyRepository) SaveMemberAccount(ctx context.Context, account *domain.MemberAccount) error {
	return r.saveMemberAccountWithTx(ctx, r.db, account)
}

func (r *loyaltyRepository) SaveMemberAccountInTx(ctx context.Context, tx any, account *domain.MemberAccount) error {
	gormTx, ok := tx.(*gorm.DB)
	if !ok || gormTx == nil {
		return errors.New("invalid transaction")
	}
	return r.saveMemberAccountWithTx(ctx, gormTx, account)
}

func (r *loyaltyRepository) GetMemberAccount(ctx context.Context, userID uint64) (*domain.MemberAccount, error) {
	var model MemberAccountModel
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toAccount(&model), nil
}

// --- 积分交易 ---

func (r *loyaltyRepository) SavePointsTransaction(ctx context.Context, transaction *domain.PointsTransaction) error {
	return r.savePointsTransactionWithTx(ctx, r.db, transaction)
}

func (r *loyaltyRepository) SavePointsTransactionInTx(ctx context.Context, tx any, transaction *domain.PointsTransaction) error {
	gormTx, ok := tx.(*gorm.DB)
	if !ok || gormTx == nil {
		return errors.New("invalid transaction")
	}
	return r.savePointsTransactionWithTx(ctx, gormTx, transaction)
}

func (r *loyaltyRepository) GetPointsTransaction(ctx context.Context, id uint64) (*domain.PointsTransaction, error) {
	var model PointsTransactionModel
	if err := r.db.WithContext(ctx).First(&model, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toTransaction(&model), nil
}

func (r *loyaltyRepository) ListPointsTransactions(ctx context.Context, userID uint64, offset, limit int) ([]*domain.PointsTransaction, int64, error) {
	var list []*PointsTransactionModel
	var total int64

	db := r.db.WithContext(ctx).Model(&PointsTransactionModel{}).Where("user_id = ?", userID)

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := db.Offset(offset).Limit(limit).Order("created_at desc").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	items := make([]*domain.PointsTransaction, len(list))
	for i, model := range list {
		items[i] = toTransaction(model)
	}

	return items, total, nil
}

// --- 会员权益 ---

func (r *loyaltyRepository) SaveMemberBenefit(ctx context.Context, benefit *domain.MemberBenefit) error {
	return r.saveMemberBenefitWithTx(ctx, r.db, benefit)
}

func (r *loyaltyRepository) SaveMemberBenefitInTx(ctx context.Context, tx any, benefit *domain.MemberBenefit) error {
	gormTx, ok := tx.(*gorm.DB)
	if !ok || gormTx == nil {
		return errors.New("invalid transaction")
	}
	return r.saveMemberBenefitWithTx(ctx, gormTx, benefit)
}

func (r *loyaltyRepository) GetMemberBenefit(ctx context.Context, id uint64) (*domain.MemberBenefit, error) {
	var model MemberBenefitModel
	if err := r.db.WithContext(ctx).First(&model, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toBenefit(&model), nil
}

func (r *loyaltyRepository) GetMemberBenefitByLevel(ctx context.Context, level domain.MemberLevel) (*domain.MemberBenefit, error) {
	var model MemberBenefitModel
	if err := r.db.WithContext(ctx).Where("level = ? AND enabled = ?", level, true).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toBenefit(&model), nil
}

func (r *loyaltyRepository) ListMemberBenefits(ctx context.Context, level domain.MemberLevel) ([]*domain.MemberBenefit, error) {
	var list []*MemberBenefitModel
	db := r.db.WithContext(ctx).Model(&MemberBenefitModel{})
	if level != "" {
		db = db.Where("level = ?", level)
	}
	if err := db.Find(&list).Error; err != nil {
		return nil, err
	}
	items := make([]*domain.MemberBenefit, len(list))
	for i, model := range list {
		items[i] = toBenefit(model)
	}
	return items, nil
}

func (r *loyaltyRepository) DeleteMemberBenefit(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&MemberBenefitModel{}, id).Error
}

func (r *loyaltyRepository) DeleteMemberBenefitInTx(ctx context.Context, tx any, id uint64) error {
	gormTx, ok := tx.(*gorm.DB)
	if !ok || gormTx == nil {
		return errors.New("invalid transaction")
	}
	return gormTx.WithContext(ctx).Delete(&MemberBenefitModel{}, id).Error
}

func (r *loyaltyRepository) saveMemberAccountWithTx(ctx context.Context, tx *gorm.DB, account *domain.MemberAccount) error {
	if account == nil {
		return nil
	}
	model := toAccountModel(account)
	if err := tx.WithContext(ctx).Save(model).Error; err != nil {
		return err
	}
	account.ID = uint64(model.ID)
	account.CreatedAt = model.CreatedAt
	account.UpdatedAt = model.UpdatedAt
	return nil
}

func (r *loyaltyRepository) savePointsTransactionWithTx(ctx context.Context, tx *gorm.DB, transaction *domain.PointsTransaction) error {
	if transaction == nil {
		return nil
	}
	model := toTransactionModel(transaction)
	if err := tx.WithContext(ctx).Save(model).Error; err != nil {
		return err
	}
	transaction.ID = uint64(model.ID)
	transaction.CreatedAt = model.CreatedAt
	transaction.UpdatedAt = model.UpdatedAt
	return nil
}

func (r *loyaltyRepository) saveMemberBenefitWithTx(ctx context.Context, tx *gorm.DB, benefit *domain.MemberBenefit) error {
	if benefit == nil {
		return nil
	}
	model := toBenefitModel(benefit)
	if err := tx.WithContext(ctx).Save(model).Error; err != nil {
		return err
	}
	benefit.ID = uint64(model.ID)
	benefit.CreatedAt = model.CreatedAt
	benefit.UpdatedAt = model.UpdatedAt
	return nil
}
