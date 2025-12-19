package application

import (
	"context"

	"github.com/wyfcoding/ecommerce/internal/coupon/domain"
)

// CouponService 结构体定义了优惠券管理相关的应用服务（外观模式）。
// 它将业务逻辑委托给 CouponManager 和 CouponQuery 处理。
type CouponService struct {
	manager *CouponManager
	query   *CouponQuery
}

// NewCouponService 创建并返回一个新的 CouponService 实例。
func NewCouponService(manager *CouponManager, query *CouponQuery) *CouponService {
	return &CouponService{
		manager: manager,
		query:   query,
	}
}

// CreateCoupon 创建一个新的优惠券模板。
func (s *CouponService) CreateCoupon(ctx context.Context, name, description string, couponType int, discountAmount, minOrderAmount int64) (*domain.Coupon, error) {
	return s.manager.CreateCoupon(ctx, name, description, couponType, discountAmount, minOrderAmount)
}

// AcquireCoupon 用户领取优惠券。
func (s *CouponService) AcquireCoupon(ctx context.Context, userID, couponID uint64) (*domain.UserCoupon, error) {
	return s.manager.AcquireCoupon(ctx, userID, couponID)
}

// UseCoupon 使用优惠券。
func (s *CouponService) UseCoupon(ctx context.Context, userCouponID uint64, userID uint64, orderID string) error {
	return s.manager.UseCoupon(ctx, userCouponID, userID, orderID)
}

// GetCoupon 获取优惠券详情。
func (s *CouponService) GetCoupon(ctx context.Context, id uint64) (*domain.Coupon, error) {
	return s.query.GetCoupon(ctx, id)
}

// ListCoupons 列出优惠券。
func (s *CouponService) ListCoupons(ctx context.Context, status int, page, pageSize int) ([]*domain.Coupon, int64, error) {
	return s.query.ListCoupons(ctx, status, page, pageSize)
}

// ListUserCoupons 获取用户的优惠券。
func (s *CouponService) ListUserCoupons(ctx context.Context, userID uint64, status string, page, pageSize int) ([]*domain.UserCoupon, int64, error) {
	return s.query.ListUserCoupons(ctx, userID, status, page, pageSize)
}

// ListActiveActivities 列出进行中的活动。
func (s *CouponService) ListActiveActivities(ctx context.Context) ([]*domain.CouponActivity, error) {
	return s.query.ListActiveActivities(ctx)
}

// CreateActivity 创建活动。
func (s *CouponService) CreateActivity(ctx context.Context, activity *domain.CouponActivity) error {
	return s.manager.CreateActivity(ctx, activity)
}
