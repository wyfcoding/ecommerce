package domain

import (
	"context"
	"time" // 导入时间包，用于查询条件。
)

// CouponRepository 是优惠券模块的仓储接口。
// 它定义了对优惠券、用户优惠券和优惠券活动实体进行数据持久化操作的契约。
// 仓储接口属于领域层，旨在将领域逻辑与数据存储的实现细节解耦。
type CouponRepository interface {
	// --- Coupon methods ---

	// SaveCoupon 将优惠券实体保存到数据存储中。
	// 如果优惠券已存在，则更新；如果不存在，则创建。
	// ctx: 上下文。
	// coupon: 待保存的优惠券实体。
	SaveCoupon(ctx context.Context, coupon *Coupon) error
	// GetCoupon 根据ID获取优惠券实体。
	GetCoupon(ctx context.Context, id uint64) (*Coupon, error)
	// GetCouponByNo 根据优惠券编号获取优惠券实体。
	GetCouponByNo(ctx context.Context, couponNo string) (*Coupon, error)
	// UpdateCoupon 更新优惠券实体的信息。
	UpdateCoupon(ctx context.Context, coupon *Coupon) error
	// DeleteCoupon 根据ID删除优惠券实体。
	DeleteCoupon(ctx context.Context, id uint64) error
	// ListCoupons 列出所有优惠券实体，支持通过状态过滤和分页。
	ListCoupons(ctx context.Context, status CouponStatus, offset, limit int) ([]*Coupon, int64, error)

	// --- UserCoupon methods ---

	// SaveUserCoupon 将用户优惠券实体保存到数据存储中。
	SaveUserCoupon(ctx context.Context, userCoupon *UserCoupon) error
	// GetUserCoupon 根据ID获取用户优惠券实体。
	GetUserCoupon(ctx context.Context, id uint64) (*UserCoupon, error)
	// ListUserCoupons 列出指定用户ID的优惠券实体，支持通过状态过滤和分页。
	ListUserCoupons(ctx context.Context, userID uint64, status string, offset, limit int) ([]*UserCoupon, int64, error)
	// UpdateUserCoupon 更新用户优惠券实体的信息。
	UpdateUserCoupon(ctx context.Context, userCoupon *UserCoupon) error

	// --- CouponActivity methods ---

	// SaveActivity 将优惠券活动实体保存到数据存储中。
	SaveActivity(ctx context.Context, activity *CouponActivity) error
	// GetActivity 根据ID获取优惠券活动实体。
	GetActivity(ctx context.Context, id uint64) (*CouponActivity, error)
	// UpdateActivity 更新优惠券活动实体的信息。
	UpdateActivity(ctx context.Context, activity *CouponActivity) error
	// ListActiveActivities 列出所有当前时间正在进行中的优惠券活动。
	ListActiveActivities(ctx context.Context, now time.Time) ([]*CouponActivity, error)
}
