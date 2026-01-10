package application

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/wyfcoding/ecommerce/internal/coupon/domain"
	"github.com/wyfcoding/pkg/algorithm"
)

// CouponManager 处理优惠券模块的写操作和核心业务流程。
type CouponManager struct {
	repo      domain.CouponRepository
	logger    *slog.Logger
	optimizer *algorithm.CouponOptimizer
}

// NewCouponManager 创建并返回一个新的 CouponManager 实例。
func NewCouponManager(repo domain.CouponRepository, logger *slog.Logger) *CouponManager {
	return &CouponManager{
		repo:      repo,
		logger:    logger,
		optimizer: algorithm.NewCouponOptimizer(),
	}
}

// CreateCoupon 创建新的优惠券模板。
func (m *CouponManager) CreateCoupon(ctx context.Context, name, description string, couponType int, discountAmount, minOrderAmount int64) (*domain.Coupon, error) {
	coupon := domain.NewCoupon(name, description, domain.CouponType(couponType), discountAmount, minOrderAmount)
	if err := m.repo.SaveCoupon(ctx, coupon); err != nil {
		m.logger.ErrorContext(ctx, "failed to create coupon", "error", err)
		return nil, err
	}
	m.logger.InfoContext(ctx, "coupon template created", "coupon_id", coupon.ID, "coupon_no", coupon.CouponNo)
	return coupon, nil
}

// AcquireCoupon 处理用户领取优惠券的逻辑，包含可用性检查与限领判定。
func (m *CouponManager) AcquireCoupon(ctx context.Context, userID, couponID uint64) (*domain.UserCoupon, error) {
	// 1. 获取优惠券模板并检查可用性
	coupon, err := m.repo.GetCoupon(ctx, couponID)
	if err != nil {
		m.logger.ErrorContext(ctx, "failed to get coupon for acquisition", "coupon_id", couponID, "error", err)
		return nil, err
	}
	if coupon == nil {
		return nil, fmt.Errorf("coupon not found")
	}

	if err := coupon.CheckAvailability(); err != nil {
		return nil, err
	}

	// 2. 判定每人限领规则
	userCoupons, total, err := m.repo.ListUserCoupons(ctx, userID, "", 0, 1000)
	if err != nil {
		return nil, err
	}
	if total > 0 {
		count := 0
		for _, uc := range userCoupons {
			if uc.CouponID == couponID {
				count++
			}
		}
		if int32(count) >= coupon.UsagePerUser {
			return nil, fmt.Errorf("user has reached the limit for this coupon")
		}
	}

	// 3. 生成并持久化用户优惠券
	userCoupon := domain.NewUserCoupon(userID, couponID, coupon.CouponNo)
	if err := m.repo.SaveUserCoupon(ctx, userCoupon); err != nil {
		m.logger.ErrorContext(ctx, "failed to save user coupon", "user_id", userID, "error", err)
		return nil, err
	}

	// 4. 原子更新模板的已发行计数
	coupon.Issue(1)
	if err := m.repo.UpdateCoupon(ctx, coupon); err != nil {
		m.logger.ErrorContext(ctx, "failed to update coupon issued count", "coupon_id", couponID, "error", err)
	}

	m.logger.InfoContext(ctx, "user acquired coupon", "user_id", userID, "coupon_id", couponID, "user_coupon_id", userCoupon.ID)
	return userCoupon, nil
}

// UseCoupon 标记优惠券为已使用状态并关联订单。
func (m *CouponManager) UseCoupon(ctx context.Context, userCouponID uint64, userID uint64, orderID string) error {
	userCoupon, err := m.repo.GetUserCoupon(ctx, userCouponID)
	if err != nil {
		m.logger.ErrorContext(ctx, "failed to get user coupon for use", "id", userCouponID, "error", err)
		return err
	}
	if userCoupon == nil {
		return fmt.Errorf("user coupon not found")
	}
	if userCoupon.UserID != userID {
		return fmt.Errorf("permission denied")
	}

	if err := userCoupon.Use(orderID); err != nil {
		return err
	}

	if err := m.repo.UpdateUserCoupon(ctx, userCoupon); err != nil {
		m.logger.ErrorContext(ctx, "failed to update user coupon state", "id", userCouponID, "error", err)
		return err
	}

	// 异步更新模板统计信息，不影响核心使用流程
	coupon, err := m.repo.GetCoupon(ctx, userCoupon.CouponID)
	if err != nil {
		m.logger.ErrorContext(ctx, "failed to get coupon for stats update", "coupon_id", userCoupon.CouponID, "error", err)
	} else if coupon != nil {
		coupon.Use()
		if err := m.repo.UpdateCoupon(ctx, coupon); err != nil {
			m.logger.ErrorContext(ctx, "failed to update coupon stats", "coupon_id", coupon.ID, "error", err)
		}
	}

	m.logger.InfoContext(ctx, "coupon used successfully", "user_id", userID, "user_coupon_id", userCouponID, "order_id", orderID)
	return nil
}

// CreateActivity 发布新的优惠券营销活动。
func (m *CouponManager) CreateActivity(ctx context.Context, activity *domain.CouponActivity) error {
	if err := m.repo.SaveActivity(ctx, activity); err != nil {
		m.logger.ErrorContext(ctx, "failed to save activity", "error", err)
		return err
	}
	m.logger.InfoContext(ctx, "coupon activity created", "activity_id", activity.ID, "name", activity.Name)
	return nil
}

// SuggestBestCoupons 调用优化算法为指定订单金额推荐最优的优惠券组合（支持叠加计算）。
func (m *CouponManager) SuggestBestCoupons(ctx context.Context, userID uint64, orderAmount int64) ([]uint64, int64, int64, error) {
	userCoupons, _, err := m.repo.ListUserCoupons(ctx, userID, "unused", 1, 100)
	if err != nil {
		return nil, 0, 0, err
	}
	if len(userCoupons) == 0 {
		return []uint64{}, orderAmount, 0, nil
	}

	algoCoupons := make([]algorithm.Coupon, 0)
	couponTemplateCache := make(map[uint64]*domain.Coupon)

	for _, uc := range userCoupons {
		template, exists := couponTemplateCache[uc.CouponID]
		if !exists {
			var err error
			template, err = m.repo.GetCoupon(ctx, uc.CouponID)
			if err != nil {
				m.logger.WarnContext(ctx, "failed to get coupon template", "coupon_id", uc.CouponID, "error", err)
				continue
			}
			if template == nil {
				continue
			}
			couponTemplateCache[uc.CouponID] = template
		}

		if err := template.CheckAvailability(); err != nil {
			continue
		}

		var algoType algorithm.CouponType
		var discountRate float64
		var reductionAmount int64

		switch template.Type {
		case domain.CouponTypeDiscount:
			algoType = algorithm.CouponTypeDiscount
			if template.DiscountAmount < 100 {
				discountRate = float64(template.DiscountAmount) / 100.0
			} else {
				discountRate = 0.9
			}
		case domain.CouponTypeCash:
			if template.MinOrderAmount > 0 {
				algoType = algorithm.CouponTypeReduction
			} else {
				algoType = algorithm.CouponTypeCash
			}
			reductionAmount = template.DiscountAmount
		default:
			continue
		}

		algoCoupons = append(algoCoupons, algorithm.Coupon{
			ID:              uint64(uc.ID),
			Type:            algoType,
			Threshold:       template.MinOrderAmount,
			DiscountRate:    discountRate,
			ReductionAmount: reductionAmount,
			MaxDiscount:     template.MaxDiscount,
			CanStack:        template.CanStack,
			Priority:        1,
		})
	}

	var bestIDs []uint64
	var finalPrice, discount int64

	if len(algoCoupons) < 20 {
		bestIDs, finalPrice, discount = m.optimizer.OptimalCombination(orderAmount, algoCoupons)
	} else {
		bestIDs, finalPrice, discount = m.optimizer.GreedyOptimization(orderAmount, algoCoupons)
	}

	m.logger.InfoContext(ctx, "coupon optimization finished", "user_id", userID, "suggested_count", len(bestIDs), "total_discount", discount)
	return bestIDs, finalPrice, discount, nil
}
