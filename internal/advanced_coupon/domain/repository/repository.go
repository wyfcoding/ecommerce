package repository

import (
	"context"
	"ecommerce/internal/advanced_coupon/domain/entity"
)

// AdvancedCouponRepository 优惠券仓储接口
type AdvancedCouponRepository interface {
	// 优惠券
	Save(ctx context.Context, coupon *entity.Coupon) error
	GetByID(ctx context.Context, id uint64) (*entity.Coupon, error)
	GetByCode(ctx context.Context, code string) (*entity.Coupon, error)
	List(ctx context.Context, status entity.CouponStatus, offset, limit int) ([]*entity.Coupon, int64, error)

	// 使用记录
	SaveUsage(ctx context.Context, usage *entity.CouponUsage) error
	CountUsageByUser(ctx context.Context, userID, couponID uint64) (int64, error)

	// 统计
	SaveStatistics(ctx context.Context, stats *entity.CouponStatistics) error
	GetStatistics(ctx context.Context, couponID uint64) (*entity.CouponStatistics, error)
}
