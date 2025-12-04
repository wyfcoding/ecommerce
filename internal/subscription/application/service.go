package application

import (
	"context"
	"errors" // 导入标准错误处理库。
	"time"   // 导入时间库。

	"github.com/wyfcoding/ecommerce/internal/subscription/domain/entity"     // 导入订阅领域的实体定义。
	"github.com/wyfcoding/ecommerce/internal/subscription/domain/repository" // 导入订阅领域的仓储接口。

	"log/slog" // 导入结构化日志库。
)

// SubscriptionService 结构体定义了订阅服务相关的应用服务。
// 它协调领域层和基础设施层，处理订阅计划的创建、用户订阅、订阅的取消和续订等业务逻辑。
type SubscriptionService struct {
	repo   repository.SubscriptionRepository // 依赖SubscriptionRepository接口，用于数据持久化操作。
	logger *slog.Logger                      // 日志记录器，用于记录服务运行时的信息和错误。
}

// NewSubscriptionService 创建并返回一个新的 SubscriptionService 实例。
func NewSubscriptionService(repo repository.SubscriptionRepository, logger *slog.Logger) *SubscriptionService {
	return &SubscriptionService{
		repo:   repo,
		logger: logger,
	}
}

// CreatePlan 创建一个新的订阅计划。
// ctx: 上下文。
// name: 计划名称。
// desc: 计划描述。
// price: 计划价格（单位：分）。
// duration: 计划持续时间（天）。
// features: 计划包含的功能列表。
// 返回创建成功的SubscriptionPlan实体和可能发生的错误。
func (s *SubscriptionService) CreatePlan(ctx context.Context, name, desc string, price uint64, duration int32, features []string) (*entity.SubscriptionPlan, error) {
	plan := &entity.SubscriptionPlan{
		Name:        name,
		Description: desc,
		Price:       price,
		Duration:    duration,
		Features:    features,
		Enabled:     true, // 新计划默认为启用状态。
	}
	// 通过仓储接口保存计划。
	if err := s.repo.SavePlan(ctx, plan); err != nil {
		s.logger.ErrorContext(ctx, "failed to create plan", "name", name, "error", err)
		return nil, err
	}
	s.logger.InfoContext(ctx, "plan created successfully", "plan_id", plan.ID, "name", name)
	return plan, nil
}

// Subscribe 订阅指定计划。
// ctx: 上下文。
// userID: 订阅用户ID。
// planID: 订阅计划ID。
// 返回创建成功的Subscription实体和可能发生的错误。
func (s *SubscriptionService) Subscribe(ctx context.Context, userID, planID uint64) (*entity.Subscription, error) {
	// 1. 检查用户是否已存在活跃订阅。
	active, err := s.repo.GetActiveSubscription(ctx, userID)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to check active subscription", "user_id", userID, "error", err)
		return nil, err
	}
	if active != nil {
		return nil, errors.New("user already has an active subscription") // 用户已有活跃订阅。
	}

	// 2. 获取订阅计划详情。
	plan, err := s.repo.GetPlan(ctx, planID)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to get plan", "plan_id", planID, "error", err)
		return nil, err
	}
	if plan == nil || !plan.Enabled {
		return nil, errors.New("plan not found or disabled") // 计划不存在或已禁用。
	}

	// 3. 创建订阅实体。
	now := time.Now()
	sub := &entity.Subscription{
		UserID:    userID,
		PlanID:    planID,
		Status:    entity.SubscriptionStatusActive,       // 初始状态为活跃。
		StartDate: now,                                   // 订阅开始时间为当前时间。
		EndDate:   now.AddDate(0, 0, int(plan.Duration)), // 订阅结束时间根据计划时长计算。
		AutoRenew: true,                                  // 默认开启自动续订。
	}

	// 4. 通过仓储接口保存订阅。
	if err := s.repo.SaveSubscription(ctx, sub); err != nil {
		s.logger.ErrorContext(ctx, "failed to save subscription", "user_id", userID, "plan_id", planID, "error", err)
		return nil, err
	}
	s.logger.InfoContext(ctx, "subscription created successfully", "subscription_id", sub.ID, "user_id", userID)
	return sub, nil
}

// Cancel 取消指定ID的订阅。
// ctx: 上下文。
// id: 订阅ID。
// 返回可能发生的错误。
func (s *SubscriptionService) Cancel(ctx context.Context, id uint64) error {
	sub, err := s.repo.GetSubscription(ctx, id)
	if err != nil {
		return err
	}
	if sub == nil {
		return errors.New("subscription not found")
	}

	// 更新订阅状态为已取消，并关闭自动续订，记录取消时间。
	sub.Status = entity.SubscriptionStatusCanceled
	sub.AutoRenew = false
	now := time.Now()
	sub.CanceledAt = &now

	// 通过仓储接口保存更新后的订阅。
	if err := s.repo.SaveSubscription(ctx, sub); err != nil {
		s.logger.ErrorContext(ctx, "failed to cancel subscription", "subscription_id", id, "error", err)
		return err
	}
	s.logger.InfoContext(ctx, "subscription canceled successfully", "subscription_id", id)
	return nil
}

// Renew 续订指定ID的订阅（手动或自动触发）。
// ctx: 上下文。
// id: 订阅ID。
// 返回可能发生的错误。
func (s *SubscriptionService) Renew(ctx context.Context, id uint64) error {
	sub, err := s.repo.GetSubscription(ctx, id)
	if err != nil {
		return err
	}
	if sub == nil {
		return errors.New("subscription not found")
	}

	// 获取订阅计划详情，以确定续订时长。
	plan, err := s.repo.GetPlan(ctx, sub.PlanID)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to get plan for renewal", "plan_id", sub.PlanID, "error", err)
		return err
	}
	if plan == nil {
		return errors.New("plan not found for renewal")
	}

	// 延长订阅结束日期。
	// 如果订阅已过期，则从当前时间开始计算新的结束日期；否则，在现有结束日期基础上延长。
	if sub.EndDate.Before(time.Now()) {
		sub.EndDate = time.Now().AddDate(0, 0, int(plan.Duration))
	} else {
		sub.EndDate = sub.EndDate.AddDate(0, 0, int(plan.Duration))
	}
	sub.Status = entity.SubscriptionStatusActive // 确保状态为活跃。

	// 通过仓储接口保存更新后的订阅。
	if err := s.repo.SaveSubscription(ctx, sub); err != nil {
		s.logger.ErrorContext(ctx, "failed to renew subscription", "subscription_id", id, "error", err)
		return err
	}
	s.logger.InfoContext(ctx, "subscription renewed successfully", "subscription_id", id)
	return nil
}

// ListPlans 获取订阅计划列表。
// ctx: 上下文。
// 返回SubscriptionPlan实体列表和可能发生的错误。
func (s *SubscriptionService) ListPlans(ctx context.Context) ([]*entity.SubscriptionPlan, error) {
	return s.repo.ListPlans(ctx, true) // 只列出启用的计划。
}

// ListSubscriptions 获取用户订阅列表。
// ctx: 上下文。
// userID: 筛选订阅的用户ID。
// page, pageSize: 分页参数。
// 返回Subscription实体列表、总数和可能发生的错误。
func (s *SubscriptionService) ListSubscriptions(ctx context.Context, userID uint64, page, pageSize int) ([]*entity.Subscription, int64, error) {
	offset := (page - 1) * pageSize
	// nil表示不按状态过滤，获取所有状态的订阅。
	return s.repo.ListSubscriptions(ctx, userID, nil, offset, pageSize)
}
