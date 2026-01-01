package persistence

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

// --- 会员账户 (MemberAccount methods) ---

// SaveMemberAccount 将会员账户实体保存到数据库。
func (r *loyaltyRepository) SaveMemberAccount(ctx context.Context, account *domain.MemberAccount) error {
	return r.db.WithContext(ctx).Save(account).Error
}

// GetMemberAccount 根据用户ID从数据库获取会员账户记录。
func (r *loyaltyRepository) GetMemberAccount(ctx context.Context, userID uint64) (*domain.MemberAccount, error) {
	var account domain.MemberAccount
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&account).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &account, nil
}

// --- 积分交易 (PointsTransaction methods) ---

// SavePointsTransaction 将积分交易实体保存到数据库。
func (r *loyaltyRepository) SavePointsTransaction(ctx context.Context, transaction *domain.PointsTransaction) error {
	return r.db.WithContext(ctx).Save(transaction).Error
}

// ListPointsTransactions 从数据库列出指定用户ID的所有积分交易记录，支持分页。
func (r *loyaltyRepository) ListPointsTransactions(ctx context.Context, userID uint64, offset, limit int) ([]*domain.PointsTransaction, int64, error) {
	var list []*domain.PointsTransaction
	var total int64

	db := r.db.WithContext(ctx).Model(&domain.PointsTransaction{}).Where("user_id = ?", userID)

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := db.Offset(offset).Limit(limit).Order("created_at desc").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// --- 会员权益 (MemberBenefit methods) ---

// SaveMemberBenefit 将会员权益实体保存到数据库。
func (r *loyaltyRepository) SaveMemberBenefit(ctx context.Context, benefit *domain.MemberBenefit) error {
	return r.db.WithContext(ctx).Save(benefit).Error
}

// GetMemberBenefit 根据ID从数据库获取会员权益记录。
func (r *loyaltyRepository) GetMemberBenefit(ctx context.Context, id uint64) (*domain.MemberBenefit, error) {
	var benefit domain.MemberBenefit
	if err := r.db.WithContext(ctx).First(&benefit, id).Error; err != nil {
		return nil, err
	}
	return &benefit, nil
}

// ListMemberBenefits 从数据库列出所有会员权益记录，支持通过会员等级过滤。
func (r *loyaltyRepository) ListMemberBenefits(ctx context.Context, level domain.MemberLevel) ([]*domain.MemberBenefit, error) {
	var list []*domain.MemberBenefit
	db := r.db.WithContext(ctx).Model(&domain.MemberBenefit{})
	if level != "" {
		db = db.Where("level = ?", level)
	}
	if err := db.Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *loyaltyRepository) GetMemberBenefitByLevel(ctx context.Context, level domain.MemberLevel) (*domain.MemberBenefit, error) {
	var benefit domain.MemberBenefit
	if err := r.db.WithContext(ctx).Where("level = ? AND enabled = ?", level, true).First(&benefit).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &benefit, nil
}

// DeleteMemberBenefit 根据ID从数据库删除会员权益记录。
func (r *loyaltyRepository) DeleteMemberBenefit(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&domain.MemberBenefit{}, id).Error
}
