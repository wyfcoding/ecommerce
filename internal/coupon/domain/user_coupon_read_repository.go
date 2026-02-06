// 生成摘要：定义用户优惠券读模型仓储接口（Redis）。
// 假设：读模型以 user_coupon_id 为主键索引。
package domain

import "context"

// UserCouponReadRepository 定义用户优惠券读模型的高性能访问接口。
type UserCouponReadRepository interface {
	// SaveUserCoupon 保存或更新用户优惠券读模型。
	SaveUserCoupon(ctx context.Context, coupon *UserCoupon) error
	// GetUserCoupon 根据用户优惠券ID获取读模型。
	GetUserCoupon(ctx context.Context, id uint64) (*UserCoupon, error)
	// DeleteUserCoupon 删除用户优惠券读模型。
	DeleteUserCoupon(ctx context.Context, id uint64) error
}
