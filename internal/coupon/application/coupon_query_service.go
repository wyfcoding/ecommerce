package application

import (
	"context"
	"log/slog"

	"github.com/wyfcoding/ecommerce/internal/coupon/domain"
)

// CouponQueryService 负责处理 Coupon 相关的读操作和查询逻辑。
type CouponQueryService struct {
	repo                 domain.CouponRepository
	couponReadRepo       domain.CouponReadRepository
	userCouponReadRepo   domain.UserCouponReadRepository
	couponSearchRepo     domain.CouponSearchRepository
	userCouponSearchRepo domain.UserCouponSearchRepository
	logger               *slog.Logger
}

// NewCouponQueryService 创建并返回一个新的 CouponQueryService 实例。
func NewCouponQueryService(
	repo domain.CouponRepository,
	couponReadRepo domain.CouponReadRepository,
	userCouponReadRepo domain.UserCouponReadRepository,
	couponSearchRepo domain.CouponSearchRepository,
	userCouponSearchRepo domain.UserCouponSearchRepository,
	logger *slog.Logger,
) *CouponQueryService {
	return &CouponQueryService{
		repo:                 repo,
		couponReadRepo:       couponReadRepo,
		userCouponReadRepo:   userCouponReadRepo,
		couponSearchRepo:     couponSearchRepo,
		userCouponSearchRepo: userCouponSearchRepo,
		logger:               logger,
	}
}

// GetCoupon 获取指定ID的优惠券模板详情。
func (q *CouponQueryService) GetCoupon(ctx context.Context, id uint64) (*domain.Coupon, error) {
	if q.couponReadRepo != nil {
		if coupon, err := q.couponReadRepo.GetCoupon(ctx, id); err == nil && coupon != nil {
			return coupon, nil
		}
	}

	coupon, err := q.repo.GetCoupon(ctx, id)
	if err != nil {
		return nil, err
	}
	if coupon != nil && q.couponReadRepo != nil {
		if err := q.couponReadRepo.SaveCoupon(ctx, coupon); err != nil {
			q.logger.WarnContext(ctx, "failed to warm coupon cache", "coupon_id", id, "error", err)
		}
	}
	return coupon, nil
}

// GetUserCoupon 获取指定ID的用户优惠券详情。
func (q *CouponQueryService) GetUserCoupon(ctx context.Context, id uint64) (*domain.UserCoupon, error) {
	if q.userCouponReadRepo != nil {
		if coupon, err := q.userCouponReadRepo.GetUserCoupon(ctx, id); err == nil && coupon != nil {
			return coupon, nil
		}
	}

	userCoupon, err := q.repo.GetUserCoupon(ctx, id)
	if err != nil {
		return nil, err
	}
	if userCoupon != nil && q.userCouponReadRepo != nil {
		if err := q.userCouponReadRepo.SaveUserCoupon(ctx, userCoupon); err != nil {
			q.logger.WarnContext(ctx, "failed to warm user coupon cache", "user_coupon_id", id, "error", err)
		}
	}
	return userCoupon, nil
}

// ListUserCoupons 获取特定用户的优惠券持有列表。
func (q *CouponQueryService) ListUserCoupons(ctx context.Context, userID uint64, status string, page, pageSize int) ([]*domain.UserCoupon, int64, error) {
	offset := (page - 1) * pageSize
	if q.userCouponSearchRepo != nil {
		list, total, err := q.userCouponSearchRepo.SearchUserCoupons(ctx, userID, status, offset, pageSize)
		if err == nil {
			return list, total, nil
		}
		q.logger.WarnContext(ctx, "user coupon search fallback to mysql", "user_id", userID, "error", err)
	}
	return q.repo.ListUserCoupons(ctx, userID, status, offset, pageSize)
}

// ListCoupons 分页获取优惠券模板列表。
func (q *CouponQueryService) ListCoupons(ctx context.Context, page, pageSize int) ([]*domain.Coupon, int64, error) {
	offset := (page - 1) * pageSize
	if q.couponSearchRepo != nil {
		list, total, err := q.couponSearchRepo.SearchCoupons(ctx, nil, offset, pageSize)
		if err == nil {
			return list, total, nil
		}
		q.logger.WarnContext(ctx, "coupon search fallback to mysql", "error", err)
	}
	return q.repo.ListCoupons(ctx, offset, pageSize)
}
