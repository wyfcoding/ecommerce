package application

import (
	"context"

	"github.com/wyfcoding/ecommerce/internal/coupon/domain"
)

// CouponQuery 负责处理 Coupon 相关的读操作和查询逻辑。
type CouponQuery struct {
	repo domain.CouponRepository
}

// NewCouponQuery 创建并返回一个新的 CouponQuery 实例。
func NewCouponQuery(repo domain.CouponRepository) *CouponQuery {
	return &CouponQuery{
		repo: repo,
	}
}

// GetCoupon 获取指定ID的优惠券模板详情。
func (q *CouponQuery) GetCoupon(ctx context.Context, id uint64) (*domain.Coupon, error) {
	return q.repo.GetCoupon(ctx, id)
}

func (q *CouponQuery) GetUserCoupon(ctx context.Context, id uint64) (*domain.UserCoupon, error) {
	return q.repo.GetUserCoupon(ctx, id)
}

func (q *CouponQuery) ListUserCoupons(ctx context.Context, userID uint64, status string, page, pageSize int) ([]*domain.UserCoupon, int64, error) {
	offset := (page - 1) * pageSize
	return q.repo.ListUserCoupons(ctx, userID, status, offset, pageSize)
}

func (q *CouponQuery) ListCoupons(ctx context.Context, page, pageSize int) ([]*domain.Coupon, int64, error) {
	offset := (page - 1) * pageSize
	return q.repo.ListCoupons(ctx, offset, pageSize)
}
