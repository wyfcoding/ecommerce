package application

import (
	"context"
	"fmt"
	"time"

	"github.com/wyfcoding/ecommerce/internal/coupon/domain/entity"     // 导入优惠券领域的实体定义。
	"github.com/wyfcoding/ecommerce/internal/coupon/domain/repository" // 导入优惠券领域的仓储接口。

	"log/slog" // 导入结构化日志库。
)

// CouponService 结构体定义了优惠券管理相关的应用服务。
// 它协调领域层和基础设施层，处理优惠券的创建、发放、使用以及优惠券活动的管理等业务逻辑。
type CouponService struct {
	repo   repository.CouponRepository // 依赖CouponRepository接口，用于数据持久化操作。
	logger *slog.Logger                // 日志记录器，用于记录服务运行时的信息和错误。
}

// NewCouponService 创建并返回一个新的 CouponService 实例。
func NewCouponService(repo repository.CouponRepository, logger *slog.Logger) *CouponService {
	return &CouponService{
		repo:   repo,
		logger: logger,
	}
}

// CreateCoupon 创建一个新的优惠券。
// ctx: 上下文。
// name: 优惠券名称。
// description: 优惠券描述。
// couponType: 优惠券类型。
// discountAmount: 优惠金额或折扣率（根据类型而定）。
// minOrderAmount: 最低订单金额门槛。
// 返回创建成功的Coupon实体和可能发生的错误。
func (s *CouponService) CreateCoupon(ctx context.Context, name, description string, couponType entity.CouponType, discountAmount, minOrderAmount int64) (*entity.Coupon, error) {
	coupon := entity.NewCoupon(name, description, couponType, discountAmount, minOrderAmount) // 创建Coupon实体。
	// 通过仓储接口保存优惠券。
	if err := s.repo.SaveCoupon(ctx, coupon); err != nil {
		s.logger.ErrorContext(ctx, "failed to create coupon", "name", name, "error", err)
		return nil, err
	}
	s.logger.InfoContext(ctx, "coupon created successfully", "coupon_id", coupon.ID, "name", name)
	return coupon, nil
}

// ActivateCoupon 激活一个指定ID的优惠券。
// ctx: 上下文。
// id: 优惠券的ID。
// 返回可能发生的错误。
func (s *CouponService) ActivateCoupon(ctx context.Context, id uint64) error {
	// 获取优惠券实体。
	coupon, err := s.repo.GetCoupon(ctx, id)
	if err != nil {
		return err
	}

	// 调用实体方法激活优惠券。
	if err := coupon.Activate(); err != nil {
		return err
	}

	// 更新数据库中的优惠券信息。
	return s.repo.UpdateCoupon(ctx, coupon)
}

// IssueCoupon 向指定用户发放一个优惠券。
// ctx: 上下文。
// userID: 目标用户ID。
// couponID: 待发放的优惠券ID。
// 返回用户获得的UserCoupon实体和可能发生的错误。
func (s *CouponService) IssueCoupon(ctx context.Context, userID, couponID uint64) (*entity.UserCoupon, error) {
	// 获取优惠券实体。
	coupon, err := s.repo.GetCoupon(ctx, couponID)
	if err != nil {
		return nil, err
	}

	// 检查优惠券是否可用。
	if err := coupon.CheckAvailability(); err != nil {
		return nil, err
	}

	// 检查用户是否已达到该优惠券的使用上限。
	// 这是一个简单的检查，生产环境可能需要更复杂的优化。
	userCoupons, _, err := s.repo.ListUserCoupons(ctx, userID, "", 0, 1000) // 获取用户所有优惠券。
	if err != nil {
		return nil, err
	}

	count := 0
	for _, uc := range userCoupons {
		if uc.CouponID == couponID {
			count++
		}
	}
	if int32(count) >= coupon.UsagePerUser { // 检查是否达到每用户使用限制。
		return nil, fmt.Errorf("user usage limit reached")
	}

	// 调用实体方法，标记优惠券已发放（减少库存）。
	coupon.Issue(1)
	if err := s.repo.UpdateCoupon(ctx, coupon); err != nil {
		return nil, err
	}

	// 创建用户优惠券实体，并保存到数据库。
	userCoupon := entity.NewUserCoupon(userID, couponID, coupon.CouponNo)
	if err := s.repo.SaveUserCoupon(ctx, userCoupon); err != nil {
		return nil, err
	}

	return userCoupon, nil
}

// UseCoupon 使用用户的优惠券。
// ctx: 上下文。
// userCouponID: 用户优惠券的ID。
// orderID: 关联的订单ID。
// 返回可能发生的错误。
func (s *CouponService) UseCoupon(ctx context.Context, userCouponID uint64, orderID string) error {
	// 获取用户优惠券实体。
	userCoupon, err := s.repo.GetUserCoupon(ctx, userCouponID)
	if err != nil {
		return err
	}

	// 调用实体方法使用优惠券。
	if err := userCoupon.Use(orderID); err != nil {
		return err
	}

	// 更新数据库中的用户优惠券信息。
	if err := s.repo.UpdateUserCoupon(ctx, userCoupon); err != nil {
		return err
	}

	// 获取优惠券实体，并标记为已使用（减少总使用次数）。
	coupon, err := s.repo.GetCoupon(ctx, userCoupon.CouponID)
	if err != nil {
		return err
	}
	coupon.Use()
	return s.repo.UpdateCoupon(ctx, coupon)
}

// ListCoupons 获取优惠券列表。
// ctx: 上下文。
// page, pageSize: 分页参数。
// 返回优惠券列表、总数和可能发生的错误。
func (s *CouponService) ListCoupons(ctx context.Context, page, pageSize int) ([]*entity.Coupon, int64, error) {
	offset := (page - 1) * pageSize
	// TODO: ListCoupons 仓储方法签名中有一个 status 参数，此处传0。
	return s.repo.ListCoupons(ctx, 0, offset, pageSize)
}

// ListUserCoupons 获取指定用户的优惠券列表。
// ctx: 上下文。
// userID: 用户ID。
// status: 优惠券状态（例如，"available", "used"）。
// page, pageSize: 分页参数。
// 返回用户优惠券列表、总数和可能发生的错误。
func (s *CouponService) ListUserCoupons(ctx context.Context, userID uint64, status string, page, pageSize int) ([]*entity.UserCoupon, int64, error) {
	offset := (page - 1) * pageSize
	return s.repo.ListUserCoupons(ctx, userID, status, offset, pageSize)
}

// CreateActivity 创建一个新的优惠券活动。
// ctx: 上下文。
// name: 活动名称。
// description: 活动描述。
// startTime, endTime: 活动的开始和结束时间。
// couponIDs: 活动中包含的优惠券ID列表。
// 返回创建成功的CouponActivity实体和可能发生的错误。
func (s *CouponService) CreateActivity(ctx context.Context, name, description string, startTime, endTime time.Time, couponIDs []uint64) (*entity.CouponActivity, error) {
	activity := entity.NewCouponActivity(name, description, startTime, endTime, couponIDs) // 创建CouponActivity实体。
	// 通过仓储接口保存活动。
	if err := s.repo.SaveActivity(ctx, activity); err != nil {
		s.logger.ErrorContext(ctx, "failed to create activity", "name", name, "error", err)
		return nil, err
	}
	s.logger.InfoContext(ctx, "activity created successfully", "activity_id", activity.ID, "name", name)
	return activity, nil
}

// ListActiveActivities 获取所有当前正在进行中的优惠券活动列表。
// ctx: 上下文。
// 返回活动列表和可能发生的错误。
func (s *CouponService) ListActiveActivities(ctx context.Context) ([]*entity.CouponActivity, error) {
	return s.repo.ListActiveActivities(ctx, time.Now())
}
