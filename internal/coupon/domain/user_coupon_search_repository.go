// 生成摘要：定义用户优惠券搜索仓储接口（Elasticsearch）。
// 假设：索引字段与 domain.UserCoupon 的 JSON 映射一致。
package domain

import "context"

// UserCouponSearchRepository 定义用户优惠券搜索的访问接口。
type UserCouponSearchRepository interface {
	// IndexUserCoupon 保存或更新用户优惠券搜索文档。
	IndexUserCoupon(ctx context.Context, coupon *UserCoupon) error
	// DeleteUserCoupon 删除用户优惠券搜索文档。
	DeleteUserCoupon(ctx context.Context, userCouponID uint64) error
	// SearchUserCoupons 分页搜索用户优惠券。
	SearchUserCoupons(ctx context.Context, userID uint64, status string, offset, limit int) ([]*UserCoupon, int64, error)
}
