package persistence

import (
	"context"
	"time"

	"github.com/wyfcoding/ecommerce/internal/coupon/domain"
	"gorm.io/gorm"
)

type couponRepository struct {
	db *gorm.DB
}

func NewCouponRepository(db *gorm.DB) domain.CouponRepository {
	return &couponRepository{db: db}
}

func (r *couponRepository) BeginTx(ctx context.Context) any {
	return r.db.WithContext(ctx).Begin()
}

func (r *couponRepository) CommitTx(tx any) error {
	return tx.(*gorm.DB).Commit().Error
}

func (r *couponRepository) RollbackTx(tx any) error {
	return tx.(*gorm.DB).Rollback().Error
}

func (r *couponRepository) SaveCoupon(ctx context.Context, coupon *domain.Coupon) error {
	return r.db.WithContext(ctx).Save(coupon).Error
}

func (r *couponRepository) SaveCouponInTx(ctx context.Context, tx any, coupon *domain.Coupon) error {
	return tx.(*gorm.DB).WithContext(ctx).Save(coupon).Error
}

func (r *couponRepository) GetCoupon(ctx context.Context, id uint64) (*domain.Coupon, error) {
	var coupon domain.Coupon
	if err := r.db.WithContext(ctx).First(&coupon, id).Error; err != nil {
		return nil, err
	}
	return &coupon, nil
}

func (r *couponRepository) GetCouponByNo(ctx context.Context, couponNo string) (*domain.Coupon, error) {
	var coupon domain.Coupon
	if err := r.db.WithContext(ctx).Where("coupon_no = ?", couponNo).First(&coupon).Error; err != nil {
		return nil, err
	}
	return &coupon, nil
}

func (r *couponRepository) UpdateCoupon(ctx context.Context, coupon *domain.Coupon) error {
	return r.db.WithContext(ctx).Save(coupon).Error
}

func (r *couponRepository) UpdateCouponInTx(ctx context.Context, tx any, coupon *domain.Coupon) error {
	return tx.(*gorm.DB).WithContext(ctx).Save(coupon).Error
}

func (r *couponRepository) DeleteCoupon(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&domain.Coupon{}, id).Error
}

func (r *couponRepository) ListCoupons(ctx context.Context, offset, limit int) ([]*domain.Coupon, int64, error) {
	var list []*domain.Coupon
	var total int64
	db := r.db.WithContext(ctx).Model(&domain.Coupon{})
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := db.Offset(offset).Limit(limit).Order("created_at desc").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *couponRepository) SaveUserCoupon(ctx context.Context, userCoupon *domain.UserCoupon) error {
	return r.db.WithContext(ctx).Save(userCoupon).Error
}

func (r *couponRepository) SaveUserCouponInTx(ctx context.Context, tx any, userCoupon *domain.UserCoupon) error {
	return tx.(*gorm.DB).WithContext(ctx).Save(userCoupon).Error
}

func (r *couponRepository) GetUserCoupon(ctx context.Context, id uint64) (*domain.UserCoupon, error) {
	var userCoupon domain.UserCoupon
	if err := r.db.WithContext(ctx).First(&userCoupon, id).Error; err != nil {
		return nil, err
	}
	return &userCoupon, nil
}

func (r *couponRepository) ListUserCoupons(ctx context.Context, userID uint64, status string, offset, limit int) ([]*domain.UserCoupon, int64, error) {
	var list []*domain.UserCoupon
	var total int64
	db := r.db.WithContext(ctx).Model(&domain.UserCoupon{}).Where("user_id = ?", userID)
	if status != "" {
		db = db.Where("status = ?", status)
	}
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := db.Offset(offset).Limit(limit).Order("created_at desc").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *couponRepository) UpdateUserCoupon(ctx context.Context, userCoupon *domain.UserCoupon) error {
	return r.db.WithContext(ctx).Save(userCoupon).Error
}

func (r *couponRepository) UpdateUserCouponInTx(ctx context.Context, tx any, userCoupon *domain.UserCoupon) error {
	return tx.(*gorm.DB).WithContext(ctx).Save(userCoupon).Error
}

func (r *couponRepository) SaveActivity(ctx context.Context, activity *domain.CouponActivity) error {
	return r.db.WithContext(ctx).Save(activity).Error
}

func (r *couponRepository) GetActivity(ctx context.Context, id uint64) (*domain.CouponActivity, error) {
	var activity domain.CouponActivity
	if err := r.db.WithContext(ctx).First(&activity, id).Error; err != nil {
		return nil, err
	}
	return &activity, nil
}

func (r *couponRepository) UpdateActivity(ctx context.Context, activity *domain.CouponActivity) error {
	return r.db.WithContext(ctx).Save(activity).Error
}

func (r *couponRepository) ListActiveActivities(ctx context.Context, now time.Time) ([]*domain.CouponActivity, error) {
	var list []*domain.CouponActivity
	if err := r.db.WithContext(ctx).Where("status = ? AND start_time <= ? AND end_time >= ?", "active", now, now).Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}
