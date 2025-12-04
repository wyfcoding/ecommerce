package persistence

import (
	"context"
	"errors" // 导入标准错误处理库。
	"time"   // 导入时间库。

	"github.com/wyfcoding/ecommerce/internal/subscription/domain/entity"     // 导入订阅领域的实体定义。
	"github.com/wyfcoding/ecommerce/internal/subscription/domain/repository" // 导入订阅领域的仓储接口。

	"gorm.io/gorm" // 导入GORM ORM框架。
)

type subscriptionRepository struct {
	db *gorm.DB // GORM数据库连接实例。
}

// NewSubscriptionRepository 创建并返回一个新的 subscriptionRepository 实例。
func NewSubscriptionRepository(db *gorm.DB) repository.SubscriptionRepository {
	return &subscriptionRepository{db: db}
}

// --- 订阅计划 (SubscriptionPlan methods) ---

// SavePlan 将订阅计划实体保存到数据库。
// 如果实体已存在，则更新；如果不存在，则创建。
func (r *subscriptionRepository) SavePlan(ctx context.Context, plan *entity.SubscriptionPlan) error {
	return r.db.WithContext(ctx).Save(plan).Error
}

// GetPlan 根据ID从数据库获取订阅计划记录。
// 如果记录未找到，则返回nil。
func (r *subscriptionRepository) GetPlan(ctx context.Context, id uint64) (*entity.SubscriptionPlan, error) {
	var plan entity.SubscriptionPlan
	if err := r.db.WithContext(ctx).First(&plan, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // 如果记录未找到，返回nil。
		}
		return nil, err // 其他错误则返回。
	}
	return &plan, nil
}

// ListPlans 从数据库列出所有订阅计划记录。
// enabledOnly: 布尔值，如果为true，则只列出启用的计划。
func (r *subscriptionRepository) ListPlans(ctx context.Context, enabledOnly bool) ([]*entity.SubscriptionPlan, error) {
	var list []*entity.SubscriptionPlan
	db := r.db.WithContext(ctx)
	if enabledOnly { // 如果enabledOnly为true，则只查询启用的计划。
		db = db.Where("enabled = ?", true)
	}
	if err := db.Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

// --- 订阅 (Subscription methods) ---

// SaveSubscription 将订阅实体保存到数据库。
func (r *subscriptionRepository) SaveSubscription(ctx context.Context, sub *entity.Subscription) error {
	return r.db.WithContext(ctx).Save(sub).Error
}

// GetSubscription 根据ID从数据库获取订阅记录。
// 如果记录未找到，则返回nil。
func (r *subscriptionRepository) GetSubscription(ctx context.Context, id uint64) (*entity.Subscription, error) {
	var sub entity.Subscription
	if err := r.db.WithContext(ctx).First(&sub, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // 如果记录未找到，返回nil。
		}
		return nil, err // 其他错误则返回。
	}
	return &sub, nil
}

// GetActiveSubscription 获取指定用户ID的活跃订阅记录。
// 活跃订阅必须状态为 Active 且未过期。
func (r *subscriptionRepository) GetActiveSubscription(ctx context.Context, userID uint64) (*entity.Subscription, error) {
	var sub entity.Subscription
	now := time.Now()
	// 查询用户ID匹配，状态为活跃，并且结束日期在当前时间之后的订阅。
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND status = ? AND end_date > ?", userID, entity.SubscriptionStatusActive, now).
		First(&sub).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // 如果记录未找到，返回nil。
		}
		return nil, err // 其他错误则返回。
	}
	return &sub, nil
}

// ListSubscriptions 从数据库列出指定用户ID的所有订阅记录，支持通过状态过滤和分页。
func (r *subscriptionRepository) ListSubscriptions(ctx context.Context, userID uint64, status *entity.SubscriptionStatus, offset, limit int) ([]*entity.Subscription, int64, error) {
	var list []*entity.Subscription
	var total int64

	db := r.db.WithContext(ctx).Model(&entity.Subscription{})
	if userID > 0 { // 如果提供了用户ID，则按用户ID过滤。
		db = db.Where("user_id = ?", userID)
	}
	if status != nil { // 如果提供了状态，则按状态过滤。
		db = db.Where("status = ?", *status)
	}

	// 统计总记录数。
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 应用分页和排序。
	if err := db.Offset(offset).Limit(limit).Order("id desc").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}
