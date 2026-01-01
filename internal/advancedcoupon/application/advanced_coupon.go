package application

import (
	"context"
	"log/slog"
	"time"

	"github.com/wyfcoding/ecommerce/internal/advancedcoupon/domain"
)

// AdvancedCouponService 高级优惠券门面服务，整合 Manager 和 Query。
type AdvancedCouponService struct {
	manager *AdvancedCouponManager
	query   *AdvancedCouponQuery
}

// NewAdvancedCouponService 构造函数。
func NewAdvancedCouponService(repo domain.AdvancedCouponRepository, logger *slog.Logger) *AdvancedCouponService {
	return &AdvancedCouponService{
		manager: NewAdvancedCouponManager(repo, logger),
		query:   NewAdvancedCouponQuery(repo, logger),
	}
}

// --- Manager (Writes) ---

func (s *AdvancedCouponService) CreateCoupon(ctx context.Context, code string, couponType domain.CouponType, discountValue int64, validFrom, validUntil time.Time, totalQuantity int64) (*domain.Coupon, error) {
	return s.manager.CreateCoupon(ctx, code, couponType, discountValue, validFrom, validUntil, totalQuantity)
}

func (s *AdvancedCouponService) UseCoupon(ctx context.Context, userID uint64, code string, orderID uint64) error {
	return s.manager.UseCoupon(ctx, userID, code, orderID)
}

// --- Query (Reads) ---

func (s *AdvancedCouponService) GetCoupon(ctx context.Context, id uint64) (*domain.Coupon, error) {
	return s.query.GetCoupon(ctx, id)
}

func (s *AdvancedCouponService) ListCoupons(ctx context.Context, status domain.CouponStatus, page, pageSize int) ([]*domain.Coupon, int64, error) {
	return s.query.ListCoupons(ctx, status, page, pageSize)
}

func (s *AdvancedCouponService) CalculateBestDiscount(ctx context.Context, orderAmount int64, couponIDs []uint64) ([]uint64, int64, int64, error) {
	return s.query.CalculateBestDiscount(ctx, orderAmount, couponIDs)
}
