package application

import (
	"context"

	"github.com/wyfcoding/ecommerce/internal/coupon/domain"
)

// Coupon 门面服务，整合 CommandService 和 Query。
type Coupon struct {
	command *CouponCommandService
	query   *CouponQuery
}

// NewCoupon 构造函数。
func NewCoupon(command *CouponCommandService, query *CouponQuery) *Coupon {
	return &Coupon{
		command: command,
		query:   query,
	}
}

// --- Commands (Writes) ---

func (s *Coupon) CreateCoupon(ctx context.Context, name, description string, couponType int, discountAmount, minOrderAmount int64) (*domain.Coupon, error) {
	return s.command.CreateCoupon(ctx, name, description, couponType, discountAmount, minOrderAmount)
}

func (s *Coupon) AcquireCoupon(ctx context.Context, userID, couponID uint64) (*domain.UserCoupon, error) {
	return s.command.AcquireCoupon(ctx, userID, couponID)
}

func (s *Coupon) UseCoupon(ctx context.Context, userCouponID uint64, userID uint64, orderID string) error {
	return s.command.UseCoupon(ctx, userCouponID, userID, orderID)
}

func (s *Coupon) CreateActivity(ctx context.Context, activity *domain.CouponActivity) error {
	return s.command.CreateActivity(ctx, activity)
}

func (s *Coupon) DeleteCoupon(ctx context.Context, id uint64) error {
	return s.command.DeleteCoupon(ctx, id)
}

// --- Query (Reads) ---

func (s *Coupon) GetCoupon(ctx context.Context, id uint64) (*domain.Coupon, error) {
	return s.query.GetCoupon(ctx, id)
}

func (s *Coupon) GetUserCoupon(ctx context.Context, id uint64) (*domain.UserCoupon, error) {
	return s.query.GetUserCoupon(ctx, id)
}

func (s *Coupon) ListUserCoupons(ctx context.Context, userID uint64, status string, page, pageSize int) ([]*domain.UserCoupon, int64, error) {
	return s.query.ListUserCoupons(ctx, userID, status, page, pageSize)
}

func (s *Coupon) ListCoupons(ctx context.Context, page, pageSize int) ([]*domain.Coupon, int64, error) {
	return s.query.ListCoupons(ctx, page, pageSize)
}

func (s *Coupon) SuggestBestCoupons(ctx context.Context, userID uint64, orderAmount int64) ([]uint64, int64, int64, error) {
	// Suggest logic can be in query or command, typically query.
	// We'll keep it in command if it was there, or move to query.
	// Since original was in CouponManager (now command), let's keep it in command for now or move to query.
	// Actually, optimization is read-intensive "logic over data", query is fine.
	return s.command.SuggestBestCoupons(ctx, userID, orderAmount)
}
