package repository

import (
	"context"
	"ecommerce/internal/coupon/domain/entity"
	"time"
)

// CouponRepository 优惠券仓储接口
type CouponRepository interface {
	// Coupon methods
	SaveCoupon(ctx context.Context, coupon *entity.Coupon) error
	GetCoupon(ctx context.Context, id uint64) (*entity.Coupon, error)
	GetCouponByNo(ctx context.Context, couponNo string) (*entity.Coupon, error)
	UpdateCoupon(ctx context.Context, coupon *entity.Coupon) error
	DeleteCoupon(ctx context.Context, id uint64) error
	ListCoupons(ctx context.Context, status entity.CouponStatus, offset, limit int) ([]*entity.Coupon, int64, error)

	// UserCoupon methods
	SaveUserCoupon(ctx context.Context, userCoupon *entity.UserCoupon) error
	GetUserCoupon(ctx context.Context, id uint64) (*entity.UserCoupon, error)
	ListUserCoupons(ctx context.Context, userID uint64, status string, offset, limit int) ([]*entity.UserCoupon, int64, error)
	UpdateUserCoupon(ctx context.Context, userCoupon *entity.UserCoupon) error

	// CouponActivity methods
	SaveActivity(ctx context.Context, activity *entity.CouponActivity) error
	GetActivity(ctx context.Context, id uint64) (*entity.CouponActivity, error)
	UpdateActivity(ctx context.Context, activity *entity.CouponActivity) error
	ListActiveActivities(ctx context.Context, now time.Time) ([]*entity.CouponActivity, error)
}
