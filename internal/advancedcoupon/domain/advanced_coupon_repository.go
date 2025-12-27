package domain

import (
	"context"
)

// AdvancedCouponRepository 优惠券仓储接口
type AdvancedCouponRepository interface {
	// 优惠券
	Save(ctx context.Context, coupon *Coupon) error
	GetByID(ctx context.Context, id uint64) (*Coupon, error)
	GetByCode(ctx context.Context, code string) (*Coupon, error)
	List(ctx context.Context, status CouponStatus, offset, limit int) ([]*Coupon, int64, error)

	// 使用记录
	SaveUsage(ctx context.Context, usage *CouponUsage) error
	CountUsageByUser(ctx context.Context, userID, couponID uint64) (int64, error)

	// 统计
	SaveStatistics(ctx context.Context, stats *CouponStatistics) error
	GetStatistics(ctx context.Context, couponID uint64) (*CouponStatistics, error)
}
