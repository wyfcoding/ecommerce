package application

import (
	"context"

	"github.com/wyfcoding/ecommerce/internal/coupon/domain"
)

// Coupon 结构体定义了优惠券管理相关的应用服务（外观模式）。
// 它将业务逻辑委托给 CouponManager 和 CouponQuery 处理。
type Coupon struct {
	manager *CouponManager
	query   *CouponQuery
}

// NewCoupon 创建并返回一个新的 Coupon 实例。
func NewCoupon(manager *CouponManager, query *CouponQuery) *Coupon {
	return &Coupon{
		manager: manager,
		query:   query,
	}
}

// CreateCoupon 创建一个新的优惠券模板。
func (s *Coupon) CreateCoupon(ctx context.Context, name, description string, couponType int, discountAmount, minOrderAmount int64) (*domain.Coupon, error) {
	return s.manager.CreateCoupon(ctx, name, description, couponType, discountAmount, minOrderAmount)
}

// AcquireCoupon 用户领取优惠券。
func (s *Coupon) AcquireCoupon(ctx context.Context, userID, couponID uint64) (*domain.UserCoupon, error) {
	return s.manager.AcquireCoupon(ctx, userID, couponID)
}

// UseCoupon 使用优惠券（核销）。
func (s *Coupon) UseCoupon(ctx context.Context, userCouponID uint64, userID uint64, orderID string) error {
	return s.manager.UseCoupon(ctx, userCouponID, userID, orderID)
}

// GetCoupon 获取指定ID的优惠券模板详情。
func (s *Coupon) GetCoupon(ctx context.Context, id uint64) (*domain.Coupon, error) {
	return s.query.GetCoupon(ctx, id)
}

// ListCoupons 列出所有优惠券模板。
func (s *Coupon) ListCoupons(ctx context.Context, status int, page, pageSize int) ([]*domain.Coupon, int64, error) {
	return s.query.ListCoupons(ctx, status, page, pageSize)
}

// ListUserCoupons 获取指定用户的所有领取后的优惠券。
func (s *Coupon) ListUserCoupons(ctx context.Context, userID uint64, status string, page, pageSize int) ([]*domain.UserCoupon, int64, error) {
	return s.query.ListUserCoupons(ctx, userID, status, page, pageSize)
}

// ListActiveActivities 列出当前所有进行中的优惠券营销活动。
func (s *Coupon) ListActiveActivities(ctx context.Context) ([]*domain.CouponActivity, error) {
	return s.query.ListActiveActivities(ctx)
}

// CreateActivity 创建一个新的优惠券营销活动。
func (s *Coupon) CreateActivity(ctx context.Context, activity *domain.CouponActivity) error {
	return s.manager.CreateActivity(ctx, activity)
}
