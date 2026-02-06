package domain

import (
	"context"
	"time"
)

// CouponRepository 是优惠券模块的仓储接口。
type CouponRepository interface {
	// 事务管理
	BeginTx(ctx context.Context) any
	CommitTx(tx any) error
	RollbackTx(tx any) error
	WithTx(ctx context.Context, fn func(tx any) error) error

	// --- Coupon methods ---
	SaveCoupon(ctx context.Context, coupon *Coupon) error
	SaveCouponInTx(ctx context.Context, tx any, coupon *Coupon) error
	GetCoupon(ctx context.Context, id uint64) (*Coupon, error)
	GetCouponByNo(ctx context.Context, couponNo string) (*Coupon, error)
	UpdateCoupon(ctx context.Context, coupon *Coupon) error
	UpdateCouponInTx(ctx context.Context, tx any, coupon *Coupon) error
	DeleteCoupon(ctx context.Context, id uint64) error
	ListCoupons(ctx context.Context, offset, limit int) ([]*Coupon, int64, error)

	// --- UserCoupon methods ---
	SaveUserCoupon(ctx context.Context, userCoupon *UserCoupon) error
	SaveUserCouponInTx(ctx context.Context, tx any, userCoupon *UserCoupon) error
	GetUserCoupon(ctx context.Context, id uint64) (*UserCoupon, error)
	ListUserCoupons(ctx context.Context, userID uint64, status string, offset, limit int) ([]*UserCoupon, int64, error)
	UpdateUserCoupon(ctx context.Context, userCoupon *UserCoupon) error
	UpdateUserCouponInTx(ctx context.Context, tx any, userCoupon *UserCoupon) error

	// --- CouponActivity methods ---
	SaveActivity(ctx context.Context, activity *CouponActivity) error
	GetActivity(ctx context.Context, id uint64) (*CouponActivity, error)
	UpdateActivity(ctx context.Context, activity *CouponActivity) error
	ListActiveActivities(ctx context.Context, now time.Time) ([]*CouponActivity, error)
}
