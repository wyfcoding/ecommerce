package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"ecommerce/internal/coupon/model"
)

// CouponRepository 定义了优惠券数据仓库的接口
type CouponRepository interface {
	// CouponDefinition 操作
	CreateDefinition(ctx context.Context, def *model.CouponDefinition) error
	GetDefinitionByCode(ctx context.Context, code string) (*model.CouponDefinition, error)
	IncrementIssuedQuantity(ctx context.Context, defID uint, count int) error

	// UserCoupon 操作
	CreateUserCoupon(ctx context.Context, userCoupon *model.UserCoupon) error
	GetUserCoupon(ctx context.Context, userID, userCouponID uint) (*model.UserCoupon, error)
	ListUserCoupons(ctx context.Context, userID uint, status model.CouponStatus) ([]model.UserCoupon, error)
	// RedeemUserCoupon 是一个关键的原子操作
	RedeemUserCoupon(ctx context.Context, userID, userCouponID uint, orderID uint) error
}

// couponRepository 是接口的具体实现
type couponRepository struct {
	db *gorm.DB
}

// NewCouponRepository 创建一个新的 couponRepository 实例
func NewCouponRepository(db *gorm.DB) CouponRepository {
	return &couponRepository{db: db}
}

// --- CouponDefinition 操作 ---

func (r *couponRepository) CreateDefinition(ctx context.Context, def *model.CouponDefinition) error {
	if err := r.db.WithContext(ctx).Create(def).Error; err != nil {
		return fmt.Errorf("数据库创建优惠券定义失败: %w", err)
	}
	return nil
}

func (r *couponRepository) GetDefinitionByCode(ctx context.Context, code string) (*model.CouponDefinition, error) {
	var def model.CouponDefinition
	if err := r.db.WithContext(ctx).Where("code = ?", code).First(&def).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("数据库查询优惠券定义失败: %w", err)
	}
	return &def, nil
}

func (r *couponRepository) IncrementIssuedQuantity(ctx context.Context, defID uint, count int) error {
	result := r.db.WithContext(ctx).Model(&model.CouponDefinition{}).
		Where("id = ?", defID).
		Update("issued_quantity", gorm.Expr("issued_quantity + ?", count))
	if result.Error != nil {
		return fmt.Errorf("数据库增加已发行数量失败: %w", result.Error)
	}
	return nil
}

// --- UserCoupon 操作 ---

func (r *couponRepository) CreateUserCoupon(ctx context.Context, userCoupon *model.UserCoupon) error {
	if err := r.db.WithContext(ctx).Create(userCoupon).Error; err != nil {
		return fmt.Errorf("数据库创建用户优惠券失败: %w", err)
	}
	return nil
}

func (r *couponRepository) GetUserCoupon(ctx context.Context, userID, userCouponID uint) (*model.UserCoupon, error) {
	var uc model.UserCoupon
	if err := r.db.WithContext(ctx).Preload("CouponDefinition").Where("id = ? AND user_id = ?", userCouponID, userID).First(&uc).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("数据库查询用户优惠券失败: %w", err)
	}
	return &uc, nil
}

func (r *couponRepository) ListUserCoupons(ctx context.Context, userID uint, status model.CouponStatus) ([]model.UserCoupon, error) {
	var coupons []model.UserCoupon
	db := r.db.WithContext(ctx).Preload("CouponDefinition").Where("user_id = ?", userID)
	if status != "" {
		db = db.Where("status = ?", status)
	}
	if err := db.Find(&coupons).Error; err != nil {
		return nil, fmt.Errorf("数据库列出用户优惠券失败: %w", err)
	}
	return coupons, nil
}

// RedeemUserCoupon 将用户优惠券标记为已使用 (原子操作)
func (r *couponRepository) RedeemUserCoupon(ctx context.Context, userID, userCouponID uint, orderID uint) error {
	// 在事务中执行，确保原子性
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var userCoupon model.UserCoupon

		// 1. 锁定要更新的行，防止并发问题
		if err := tx.Set("gorm:query_option", "FOR UPDATE").Where("id = ? AND user_id = ?", userCouponID, userID).First(&userCoupon).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return fmt.Errorf("优惠券不存在或不属于该用户")
			}
			return err
		}

		// 2. 检查优惠券状态是否为 UNUSED
		if userCoupon.Status != model.StatusUnused {
			return fmt.Errorf("优惠券已被使用或已过期")
		}

		// 3. 更新状态和核销信息
		updates := map[string]interface{}{
			"status":      model.StatusUsed,
			"redeemed_at": gorm.Expr("NOW()"),
			"order_id":    orderID,
		}

		if err := tx.Model(&userCoupon).Updates(updates).Error; err != nil {
			return err
		}

		return nil
	})
}
