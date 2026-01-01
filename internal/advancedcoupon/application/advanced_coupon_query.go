package application

import (
	"context"
	"log/slog"

	"github.com/wyfcoding/ecommerce/internal/advancedcoupon/domain"
	"github.com/wyfcoding/pkg/algorithm"
)

// AdvancedCouponQuery 处理所有高级优惠券相关的查询操作（Queries）。
type AdvancedCouponQuery struct {
	repo   domain.AdvancedCouponRepository
	logger *slog.Logger
}

// NewAdvancedCouponQuery 构造函数。
func NewAdvancedCouponQuery(repo domain.AdvancedCouponRepository, logger *slog.Logger) *AdvancedCouponQuery {
	return &AdvancedCouponQuery{
		repo:   repo,
		logger: logger,
	}
}

// GetCoupon 获取指定ID的优惠券详情。
func (q *AdvancedCouponQuery) GetCoupon(ctx context.Context, id uint64) (*domain.Coupon, error) {
	return q.repo.GetByID(ctx, id)
}

// ListCoupons 根据状态分页列出优惠券模板。
func (q *AdvancedCouponQuery) ListCoupons(ctx context.Context, status domain.CouponStatus, page, pageSize int) ([]*domain.Coupon, int64, error) {
	offset := (page - 1) * pageSize
	return q.repo.List(ctx, status, offset, pageSize)
}

// CalculateBestDiscount 核心算法：基于多种约束计算订单的最优优惠组合及最终金额。
func (q *AdvancedCouponQuery) CalculateBestDiscount(ctx context.Context, orderAmount int64, couponIDs []uint64) ([]uint64, int64, int64, error) {
	if len(couponIDs) == 0 {
		return nil, orderAmount, 0, nil
	}

	var algoCoupons []algorithm.Coupon
	for _, id := range couponIDs {
		coupon, err := q.repo.GetByID(ctx, id)
		if err != nil {
			continue
		}
		if coupon == nil || !coupon.IsValid() {
			continue
		}

		ac := algorithm.Coupon{
			ID:        uint64(coupon.ID),
			Threshold: coupon.MinPurchaseAmount,
			CanStack:  true,
			Priority:  1,
		}

		switch coupon.Type {
		case domain.CouponTypePercentage:
			ac.Type = algorithm.CouponTypeDiscount
			ac.DiscountRate = 1.0 - float64(coupon.DiscountValue)/100.0
			ac.MaxDiscount = coupon.MaxDiscountAmount
		case domain.CouponTypeFixed:
			ac.Type = algorithm.CouponTypeReduction
			ac.ReductionAmount = coupon.DiscountValue
		case domain.CouponTypeFreeShipping:
			continue
		}

		algoCoupons = append(algoCoupons, ac)
	}

	optimizer := algorithm.NewCouponOptimizer()
	bestCombination, finalPrice, discount := optimizer.OptimalCombination(orderAmount, algoCoupons)

	return bestCombination, finalPrice, discount, nil
}
