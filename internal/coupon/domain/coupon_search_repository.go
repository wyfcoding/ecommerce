// 生成摘要：定义优惠券搜索仓储接口（Elasticsearch）。
// 假设：索引字段与 domain.Coupon 的 JSON 映射一致。
package domain

import "context"

// CouponSearchRepository 定义优惠券搜索的访问接口。
type CouponSearchRepository interface {
	// IndexCoupon 保存或更新优惠券搜索文档。
	IndexCoupon(ctx context.Context, coupon *Coupon) error
	// DeleteCoupon 删除优惠券搜索文档。
	DeleteCoupon(ctx context.Context, couponID uint64) error
	// SearchCoupons 分页搜索优惠券。
	SearchCoupons(ctx context.Context, status *CouponStatus, offset, limit int) ([]*Coupon, int64, error)
}
