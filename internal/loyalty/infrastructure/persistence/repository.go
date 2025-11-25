package persistence

import (
	"context"
	"github.com/wyfcoding/ecommerce/internal/loyalty/domain/entity"
	"github.com/wyfcoding/ecommerce/internal/loyalty/domain/repository"
	"errors"

	"gorm.io/gorm"
)

type loyaltyRepository struct {
	db *gorm.DB
}

func NewLoyaltyRepository(db *gorm.DB) repository.LoyaltyRepository {
	return &loyaltyRepository{db: db}
}

// 会员账户
func (r *loyaltyRepository) SaveMemberAccount(ctx context.Context, account *entity.MemberAccount) error {
	return r.db.WithContext(ctx).Save(account).Error
}

func (r *loyaltyRepository) GetMemberAccount(ctx context.Context, userID uint64) (*entity.MemberAccount, error) {
	var account entity.MemberAccount
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&account).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // Return nil if not found, let service handle creation
		}
		return nil, err
	}
	return &account, nil
}

// 积分交易
func (r *loyaltyRepository) SavePointsTransaction(ctx context.Context, transaction *entity.PointsTransaction) error {
	return r.db.WithContext(ctx).Save(transaction).Error
}

func (r *loyaltyRepository) ListPointsTransactions(ctx context.Context, userID uint64, offset, limit int) ([]*entity.PointsTransaction, int64, error) {
	var list []*entity.PointsTransaction
	var total int64

	db := r.db.WithContext(ctx).Model(&entity.PointsTransaction{}).Where("user_id = ?", userID)

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := db.Offset(offset).Limit(limit).Order("created_at desc").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// 会员权益
func (r *loyaltyRepository) SaveMemberBenefit(ctx context.Context, benefit *entity.MemberBenefit) error {
	return r.db.WithContext(ctx).Save(benefit).Error
}

func (r *loyaltyRepository) GetMemberBenefit(ctx context.Context, id uint64) (*entity.MemberBenefit, error) {
	var benefit entity.MemberBenefit
	if err := r.db.WithContext(ctx).First(&benefit, id).Error; err != nil {
		return nil, err
	}
	return &benefit, nil
}

func (r *loyaltyRepository) ListMemberBenefits(ctx context.Context, level entity.MemberLevel) ([]*entity.MemberBenefit, error) {
	var list []*entity.MemberBenefit
	db := r.db.WithContext(ctx).Model(&entity.MemberBenefit{})
	if level != "" {
		db = db.Where("level = ?", level)
	}
	if err := db.Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *loyaltyRepository) DeleteMemberBenefit(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&entity.MemberBenefit{}, id).Error
}
