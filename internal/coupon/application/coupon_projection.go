// 生成摘要：新增优惠券读模型投影服务，消费事件后刷新 Redis/ES 读侧。
// 假设：读模型以 coupon_id 与 user_coupon_id 为主键，写模型为最终一致性来源。
package application

import (
	"context"
	"log/slog"

	"github.com/wyfcoding/ecommerce/internal/coupon/domain"
)

// CouponProjectionService 负责将事件转换为读模型更新。
type CouponProjectionService struct {
	repo                 domain.CouponRepository
	couponReadRepo       domain.CouponReadRepository
	userCouponReadRepo   domain.UserCouponReadRepository
	couponSearchRepo     domain.CouponSearchRepository
	userCouponSearchRepo domain.UserCouponSearchRepository
	logger               *slog.Logger
}

// NewCouponProjectionService 创建优惠券投影服务。
func NewCouponProjectionService(
	repo domain.CouponRepository,
	couponReadRepo domain.CouponReadRepository,
	userCouponReadRepo domain.UserCouponReadRepository,
	couponSearchRepo domain.CouponSearchRepository,
	userCouponSearchRepo domain.UserCouponSearchRepository,
	logger *slog.Logger,
) *CouponProjectionService {
	return &CouponProjectionService{
		repo:                 repo,
		couponReadRepo:       couponReadRepo,
		userCouponReadRepo:   userCouponReadRepo,
		couponSearchRepo:     couponSearchRepo,
		userCouponSearchRepo: userCouponSearchRepo,
		logger:               logger,
	}
}

// OnCouponCreated 处理优惠券创建事件。
func (s *CouponProjectionService) OnCouponCreated(ctx context.Context, event *domain.CouponCreatedEvent) error {
	if event == nil {
		return nil
	}
	return s.refreshCoupon(ctx, event.CouponID)
}

// OnCouponUpdated 处理优惠券更新事件。
func (s *CouponProjectionService) OnCouponUpdated(ctx context.Context, event *domain.CouponUpdatedEvent) error {
	if event == nil {
		return nil
	}
	return s.refreshCoupon(ctx, event.CouponID)
}

// OnCouponDeleted 处理优惠券删除事件。
func (s *CouponProjectionService) OnCouponDeleted(ctx context.Context, event *domain.CouponDeletedEvent) error {
	if event == nil {
		return nil
	}
	return s.refreshCoupon(ctx, event.CouponID)
}

// OnCouponIssued 处理优惠券发放事件。
func (s *CouponProjectionService) OnCouponIssued(ctx context.Context, event *domain.CouponIssuedEvent) error {
	if event == nil {
		return nil
	}
	if err := s.refreshCoupon(ctx, event.CouponID); err != nil {
		return err
	}
	return s.refreshUserCoupon(ctx, event.UserCouponID)
}

// OnCouponUsed 处理优惠券核销事件。
func (s *CouponProjectionService) OnCouponUsed(ctx context.Context, event *domain.CouponUsedEvent) error {
	if event == nil {
		return nil
	}
	return s.refreshUserCoupon(ctx, event.UserCouponID)
}

// OnCouponExpired 处理优惠券过期事件。
func (s *CouponProjectionService) OnCouponExpired(ctx context.Context, event *domain.CouponExpiredEvent) error {
	if event == nil {
		return nil
	}
	return s.refreshUserCoupon(ctx, event.UserCouponID)
}

func (s *CouponProjectionService) refreshCoupon(ctx context.Context, couponID uint64) error {
	if couponID == 0 {
		return nil
	}
	coupon, err := s.repo.GetCoupon(ctx, couponID)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to load coupon for projection", "coupon_id", couponID, "error", err)
		return err
	}
	if coupon == nil {
		if s.couponReadRepo != nil {
			_ = s.couponReadRepo.DeleteCoupon(ctx, couponID, "")
		}
		if s.couponSearchRepo != nil {
			_ = s.couponSearchRepo.DeleteCoupon(ctx, couponID)
		}
		return nil
	}
	if s.couponReadRepo != nil {
		if err := s.couponReadRepo.SaveCoupon(ctx, coupon); err != nil {
			s.logger.ErrorContext(ctx, "failed to save coupon read model", "coupon_id", couponID, "error", err)
			return err
		}
	}
	if s.couponSearchRepo != nil {
		if err := s.couponSearchRepo.IndexCoupon(ctx, coupon); err != nil {
			s.logger.ErrorContext(ctx, "failed to index coupon", "coupon_id", couponID, "error", err)
			return err
		}
	}
	return nil
}

func (s *CouponProjectionService) refreshUserCoupon(ctx context.Context, userCouponID uint64) error {
	if userCouponID == 0 {
		return nil
	}
	userCoupon, err := s.repo.GetUserCoupon(ctx, userCouponID)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to load user coupon for projection", "user_coupon_id", userCouponID, "error", err)
		return err
	}
	if userCoupon == nil {
		if s.userCouponReadRepo != nil {
			_ = s.userCouponReadRepo.DeleteUserCoupon(ctx, userCouponID)
		}
		if s.userCouponSearchRepo != nil {
			_ = s.userCouponSearchRepo.DeleteUserCoupon(ctx, userCouponID)
		}
		return nil
	}
	if s.userCouponReadRepo != nil {
		if err := s.userCouponReadRepo.SaveUserCoupon(ctx, userCoupon); err != nil {
			s.logger.ErrorContext(ctx, "failed to save user coupon read model", "user_coupon_id", userCouponID, "error", err)
			return err
		}
	}
	if s.userCouponSearchRepo != nil {
		if err := s.userCouponSearchRepo.IndexUserCoupon(ctx, userCoupon); err != nil {
			s.logger.ErrorContext(ctx, "failed to index user coupon", "user_coupon_id", userCouponID, "error", err)
			return err
		}
	}
	return nil
}
