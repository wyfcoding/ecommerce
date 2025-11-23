package persistence

import (
	"context"
	"ecommerce/internal/coupon/domain/entity"
	"ecommerce/internal/coupon/domain/repository"
	"time"

	"gorm.io/gorm"
)

type couponRepository struct {
	db *gorm.DB
}

func NewCouponRepository(db *gorm.DB) repository.CouponRepository {
	return &couponRepository{db: db}
}

// Coupon methods
func (r *couponRepository) SaveCoupon(ctx context.Context, coupon *entity.Coupon) error {
	return r.db.WithContext(ctx).Save(coupon).Error
}

func (r *couponRepository) GetCoupon(ctx context.Context, id uint64) (*entity.Coupon, error) {
	var coupon entity.Coupon
	if err := r.db.WithContext(ctx).First(&coupon, id).Error; err != nil {
		return nil, err
	}
	return &coupon, nil
}

func (r *couponRepository) GetCouponByNo(ctx context.Context, couponNo string) (*entity.Coupon, error) {
	var coupon entity.Coupon
	if err := r.db.WithContext(ctx).Where("coupon_no = ?", couponNo).First(&coupon).Error; err != nil {
		return nil, err
	}
	return &coupon, nil
}

func (r *couponRepository) UpdateCoupon(ctx context.Context, coupon *entity.Coupon) error {
	return r.db.WithContext(ctx).Save(coupon).Error
}

func (r *couponRepository) DeleteCoupon(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&entity.Coupon{}, id).Error
}

func (r *couponRepository) ListCoupons(ctx context.Context, status entity.CouponStatus, offset, limit int) ([]*entity.Coupon, int64, error) {
	var list []*entity.Coupon
	var total int64

	db := r.db.WithContext(ctx).Model(&entity.Coupon{})
	if status != 0 {
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

// UserCoupon methods
func (r *couponRepository) SaveUserCoupon(ctx context.Context, userCoupon *entity.UserCoupon) error {
	return r.db.WithContext(ctx).Save(userCoupon).Error
}

func (r *couponRepository) GetUserCoupon(ctx context.Context, id uint64) (*entity.UserCoupon, error) {
	var userCoupon entity.UserCoupon
	if err := r.db.WithContext(ctx).First(&userCoupon, id).Error; err != nil {
		return nil, err
	}
	return &userCoupon, nil
}

func (r *couponRepository) ListUserCoupons(ctx context.Context, userID uint64, status string, offset, limit int) ([]*entity.UserCoupon, int64, error) {
	var list []*entity.UserCoupon
	var total int64

	db := r.db.WithContext(ctx).Model(&entity.UserCoupon{}).Where("user_id = ?", userID)
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

func (r *couponRepository) UpdateUserCoupon(ctx context.Context, userCoupon *entity.UserCoupon) error {
	return r.db.WithContext(ctx).Save(userCoupon).Error
}

// CouponActivity methods
func (r *couponRepository) SaveActivity(ctx context.Context, activity *entity.CouponActivity) error {
	return r.db.WithContext(ctx).Save(activity).Error
}

func (r *couponRepository) GetActivity(ctx context.Context, id uint64) (*entity.CouponActivity, error) {
	var activity entity.CouponActivity
	if err := r.db.WithContext(ctx).First(&activity, id).Error; err != nil {
		return nil, err
	}
	return &activity, nil
}

func (r *couponRepository) UpdateActivity(ctx context.Context, activity *entity.CouponActivity) error {
	return r.db.WithContext(ctx).Save(activity).Error
}

func (r *couponRepository) ListActiveActivities(ctx context.Context, now time.Time) ([]*entity.CouponActivity, error) {
	var list []*entity.CouponActivity
	if err := r.db.WithContext(ctx).Where("status = ? AND start_time <= ? AND end_time >= ?", "active", now, now).Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}
