// 生成摘要：定义优惠券读模型仓储接口（Redis），用于高频查询。
// 假设：读模型以 coupon_id 与 coupon_no 为主键索引。
package domain

import "context"

// CouponReadRepository 定义优惠券读模型的高性能访问接口。
type CouponReadRepository interface {
	// SaveCoupon 保存或更新优惠券读模型。
	SaveCoupon(ctx context.Context, coupon *Coupon) error
	// GetCoupon 根据优惠券ID获取读模型。
	GetCoupon(ctx context.Context, id uint64) (*Coupon, error)
	// GetCouponByNo 根据优惠券编号获取读模型。
	GetCouponByNo(ctx context.Context, couponNo string) (*Coupon, error)
	// DeleteCoupon 删除读模型数据。
	DeleteCoupon(ctx context.Context, id uint64, couponNo string) error
}
