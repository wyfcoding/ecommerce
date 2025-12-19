package application

import (
	"context"
	"time"

	"github.com/wyfcoding/ecommerce/internal/coupon/domain"
)

// CouponQuery 处理优惠券模块的查询操作。
type CouponQuery struct {
	repo domain.CouponRepository
}

// NewCouponQuery 创建并返回一个新的 CouponQuery 实例。
func NewCouponQuery(repo domain.CouponRepository) *CouponQuery {
	return &CouponQuery{repo: repo}
}

// GetCoupon 获取指定ID的优惠券模板详情。
func (q *CouponQuery) GetCoupon(ctx context.Context, id uint64) (*domain.Coupon, error) {
	return q.repo.GetCoupon(ctx, id)
}

// ListCoupons 列出优惠券模板。
func (q *CouponQuery) ListCoupons(ctx context.Context, status int, page, pageSize int) ([]*domain.Coupon, int64, error) {
	offset := (page - 1) * pageSize
	return q.repo.ListCoupons(ctx, domain.CouponStatus(status), offset, pageSize)
}

// ListUserCoupons 获取指定用户的优惠券列表。
func (q *CouponQuery) ListUserCoupons(ctx context.Context, userID uint64, status string, page, pageSize int) ([]*domain.UserCoupon, int64, error) {
	offset := (page - 1) * pageSize
	return q.repo.ListUserCoupons(ctx, userID, status, offset, pageSize)
}

// ListActiveActivities 获取当前正在进行的活动。
func (q *CouponQuery) ListActiveActivities(ctx context.Context) ([]*domain.CouponActivity, error) {
	return q.repo.ListActiveActivities(ctx, time.Now())
}
