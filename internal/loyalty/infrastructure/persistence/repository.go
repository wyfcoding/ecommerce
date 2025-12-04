package persistence

import (
	"context"
	"errors"
	"github.com/wyfcoding/ecommerce/internal/loyalty/domain/entity"     // 导入忠诚度模块的领域实体定义。
	"github.com/wyfcoding/ecommerce/internal/loyalty/domain/repository" // 导入忠诚度模块的领域仓储接口。

	"gorm.io/gorm" // 导入GORM ORM框架。
)

type loyaltyRepository struct {
	db *gorm.DB // GORM数据库连接实例。
}

// NewLoyaltyRepository 创建并返回一个新的 loyaltyRepository 实例。
// db: GORM数据库连接实例。
func NewLoyaltyRepository(db *gorm.DB) repository.LoyaltyRepository {
	return &loyaltyRepository{db: db}
}

// --- 会员账户 (MemberAccount methods) ---

// SaveMemberAccount 将会员账户实体保存到数据库。
// 如果账户已存在（通过UserID），则更新其信息；如果不存在，则创建。
func (r *loyaltyRepository) SaveMemberAccount(ctx context.Context, account *entity.MemberAccount) error {
	return r.db.WithContext(ctx).Save(account).Error
}

// GetMemberAccount 根据用户ID从数据库获取会员账户记录。
// 如果记录未找到，则返回nil而非错误，由应用层进行判断。
func (r *loyaltyRepository) GetMemberAccount(ctx context.Context, userID uint64) (*entity.MemberAccount, error) {
	var account entity.MemberAccount
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&account).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // 如果记录未找到，返回nil。
		}
		return nil, err // 其他错误则返回。
	}
	return &account, nil
}

// --- 积分交易 (PointsTransaction methods) ---

// SavePointsTransaction 将积分交易实体保存到数据库。
func (r *loyaltyRepository) SavePointsTransaction(ctx context.Context, transaction *entity.PointsTransaction) error {
	return r.db.WithContext(ctx).Save(transaction).Error
}

// ListPointsTransactions 从数据库列出指定用户ID的所有积分交易记录，支持分页。
func (r *loyaltyRepository) ListPointsTransactions(ctx context.Context, userID uint64, offset, limit int) ([]*entity.PointsTransaction, int64, error) {
	var list []*entity.PointsTransaction
	var total int64

	db := r.db.WithContext(ctx).Model(&entity.PointsTransaction{}).Where("user_id = ?", userID)

	// 统计总记录数。
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 应用分页和排序。
	if err := db.Offset(offset).Limit(limit).Order("created_at desc").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// --- 会员权益 (MemberBenefit methods) ---

// SaveMemberBenefit 将会员权益实体保存到数据库。
func (r *loyaltyRepository) SaveMemberBenefit(ctx context.Context, benefit *entity.MemberBenefit) error {
	return r.db.WithContext(ctx).Save(benefit).Error
}

// GetMemberBenefit 根据ID从数据库获取会员权益记录。
func (r *loyaltyRepository) GetMemberBenefit(ctx context.Context, id uint64) (*entity.MemberBenefit, error) {
	var benefit entity.MemberBenefit
	if err := r.db.WithContext(ctx).First(&benefit, id).Error; err != nil {
		return nil, err
	}
	return &benefit, nil
}

// ListMemberBenefits 从数据库列出所有会员权益记录，支持通过会员等级过滤。
func (r *loyaltyRepository) ListMemberBenefits(ctx context.Context, level entity.MemberLevel) ([]*entity.MemberBenefit, error) {
	var list []*entity.MemberBenefit
	db := r.db.WithContext(ctx).Model(&entity.MemberBenefit{})
	if level != "" { // 如果提供了会员等级，则按等级过滤。
		db = db.Where("level = ?", level)
	}
	if err := db.Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

// DeleteMemberBenefit 根据ID从数据库删除会员权益记录。
func (r *loyaltyRepository) DeleteMemberBenefit(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&entity.MemberBenefit{}, id).Error
}
