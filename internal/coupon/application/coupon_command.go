package application

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/wyfcoding/ecommerce/internal/coupon/domain"
	algorithm "github.com/wyfcoding/pkg/algorithm/optimization"
	"github.com/wyfcoding/pkg/messagequeue"
)

// CouponCommandService 处理优惠券模块的写操作和核心业务流程。
type CouponCommandService struct {
	repo      domain.CouponRepository
	publisher messagequeue.EventPublisher
	logger    *slog.Logger
	optimizer *algorithm.CouponOptimizer
}

// NewCouponCommandService 创建并返回一个新的 CouponCommandService 实例。
func NewCouponCommandService(repo domain.CouponRepository, publisher messagequeue.EventPublisher, logger *slog.Logger) *CouponCommandService {
	return &CouponCommandService{
		repo:      repo,
		publisher: publisher,
		logger:    logger,
		optimizer: algorithm.NewCouponOptimizer(),
	}
}

// CreateCoupon 创建新的优惠券模板。
func (m *CouponCommandService) CreateCoupon(ctx context.Context, name, description string, couponType int, discountAmount, minOrderAmount int64) (*domain.Coupon, error) {
	coupon := domain.NewCoupon(name, description, domain.CouponType(couponType), discountAmount, minOrderAmount)
	if err := m.repo.WithTx(ctx, func(tx any) error {
		if err := m.repo.SaveCouponInTx(ctx, tx, coupon); err != nil {
			m.logger.ErrorContext(ctx, "failed to create coupon", "error", err)
			return err
		}
		if m.publisher != nil {
			event := &domain.CouponCreatedEvent{
				CouponID:       coupon.ID,
				CouponNo:       coupon.CouponNo,
				Name:           name,
				DiscountAmount: discountAmount,
				Timestamp:      time.Now(),
			}
			if err := m.publisher.PublishInTx(ctx, tx, domain.CouponCreatedEventType, coupon.CouponNo, event); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}

	m.logger.InfoContext(ctx, "coupon template created", "coupon_id", coupon.ID, "coupon_no", coupon.CouponNo)
	return coupon, nil
}

// AcquireCoupon 处理用户领取优惠券的逻辑。
func (m *CouponCommandService) AcquireCoupon(ctx context.Context, userID, couponID uint64) (*domain.UserCoupon, error) {
	coupon, err := m.repo.GetCoupon(ctx, couponID)
	if err != nil {
		return nil, err
	}
	if coupon == nil {
		return nil, fmt.Errorf("coupon not found")
	}

	if err := coupon.CheckAvailability(); err != nil {
		return nil, err
	}

	// 检查限领（应用层检查，保持原有逻辑）
	userCoupons, total, err := m.repo.ListUserCoupons(ctx, userID, "", 0, 1000)
	if err == nil && total > 0 {
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

	userCoupon := domain.NewUserCoupon(userID, couponID, coupon.CouponNo)
	if err := m.repo.WithTx(ctx, func(tx any) error {
		if err := m.repo.SaveUserCouponInTx(ctx, tx, userCoupon); err != nil {
			return err
		}

		coupon.Issue(1)
		if err := m.repo.UpdateCouponInTx(ctx, tx, coupon); err != nil {
			return err
		}

		if m.publisher != nil {
			event := &domain.CouponIssuedEvent{
				UserCouponID: userCoupon.ID,
				UserID:       userID,
				CouponID:     couponID,
				CouponNo:     coupon.CouponNo,
				Timestamp:    time.Now(),
			}
			if err := m.publisher.PublishInTx(ctx, tx, domain.CouponIssuedEventType, fmt.Sprintf("%d", userCoupon.ID), event); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}

	m.logger.InfoContext(ctx, "user acquired coupon", "user_id", userID, "coupon_id", couponID, "user_coupon_id", userCoupon.ID)
	return userCoupon, nil
}

// UseCoupon 标记优惠券为已使用状态并关联订单。
func (m *CouponCommandService) UseCoupon(ctx context.Context, userCouponID uint64, userID uint64, orderID string) error {
	userCoupon, err := m.repo.GetUserCoupon(ctx, userCouponID)
	if err != nil {
		return err
	}
	if userCoupon == nil || userCoupon.UserID != userID {
		return fmt.Errorf("user coupon not found or permission denied")
	}

	return m.repo.WithTx(ctx, func(tx any) error {
		if err := userCoupon.Use(orderID); err != nil {
			return err
		}
		if err := m.repo.UpdateUserCouponInTx(ctx, tx, userCoupon); err != nil {
			return err
		}

		if m.publisher != nil {
			event := &domain.CouponUsedEvent{
				UserCouponID: userCouponID,
				UserID:       userID,
				OrderID:      orderID,
				Timestamp:    time.Now(),
			}
			if err := m.publisher.PublishInTx(ctx, tx, domain.CouponUsedEventType, fmt.Sprintf("%d", userCouponID), event); err != nil {
				return err
			}
		}
		return nil
	})
}

func (m *CouponCommandService) CreateActivity(ctx context.Context, activity *domain.CouponActivity) error {
	return m.repo.SaveActivity(ctx, activity)
}

// SuggestBestCoupons 调用优化算法为指定订单金额推荐最优的优惠券组合。
func (m *CouponCommandService) SuggestBestCoupons(ctx context.Context, userID uint64, orderAmount int64) ([]uint64, int64, int64, error) {
	userCoupons, _, err := m.repo.ListUserCoupons(ctx, userID, "unused", 0, 100)
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

func (m *CouponCommandService) DeleteCoupon(ctx context.Context, id uint64) error {
	return m.repo.DeleteCoupon(ctx, id)
}
