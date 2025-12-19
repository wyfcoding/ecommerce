package persistence

import (
	"context"
	"errors"
	"time"

	"github.com/wyfcoding/ecommerce/internal/subscription/domain"

	"gorm.io/gorm"
)

type subscriptionRepository struct {
	db *gorm.DB
}

// NewSubscriptionRepository 创建并返回一个新的 subscriptionRepository 实例。
func NewSubscriptionRepository(db *gorm.DB) domain.SubscriptionRepository {
	return &subscriptionRepository{db: db}
}

// --- 订阅计划 (SubscriptionPlan methods) ---

// SavePlan 将订阅计划实体保存到数据库。
func (r *subscriptionRepository) SavePlan(ctx context.Context, plan *domain.SubscriptionPlan) error {
	return r.db.WithContext(ctx).Save(plan).Error
}

// GetPlan 根据ID从数据库获取订阅计划记录。
func (r *subscriptionRepository) GetPlan(ctx context.Context, id uint64) (*domain.SubscriptionPlan, error) {
	var plan domain.SubscriptionPlan
	if err := r.db.WithContext(ctx).First(&plan, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &plan, nil
}

// ListPlans 从数据库列出所有订阅计划记录。
func (r *subscriptionRepository) ListPlans(ctx context.Context, enabledOnly bool) ([]*domain.SubscriptionPlan, error) {
	var list []*domain.SubscriptionPlan
	db := r.db.WithContext(ctx)
	if enabledOnly {
		db = db.Where("enabled = ?", true)
	}
	if err := db.Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

// --- 订阅 (Subscription methods) ---

// SaveSubscription 将订阅实体保存到数据库。
func (r *subscriptionRepository) SaveSubscription(ctx context.Context, sub *domain.Subscription) error {
	return r.db.WithContext(ctx).Save(sub).Error
}

// GetSubscription 根据ID从数据库获取订阅记录。
func (r *subscriptionRepository) GetSubscription(ctx context.Context, id uint64) (*domain.Subscription, error) {
	var sub domain.Subscription
	if err := r.db.WithContext(ctx).First(&sub, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &sub, nil
}

// GetActiveSubscription 获取指定用户ID的活跃订阅记录。
func (r *subscriptionRepository) GetActiveSubscription(ctx context.Context, userID uint64) (*domain.Subscription, error) {
	var sub domain.Subscription
	now := time.Now()
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND status = ? AND end_date > ?", userID, domain.SubscriptionStatusActive, now).
		First(&sub).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &sub, nil
}

// ListSubscriptions 从数据库列出指定用户ID的所有订阅记录，支持通过状态过滤和分页。
func (r *subscriptionRepository) ListSubscriptions(ctx context.Context, userID uint64, status *domain.SubscriptionStatus, offset, limit int) ([]*domain.Subscription, int64, error) {
	var list []*domain.Subscription
	var total int64

	db := r.db.WithContext(ctx).Model(&domain.Subscription{})
	if userID > 0 {
		db = db.Where("user_id = ?", userID)
	}
	if status != nil {
		db = db.Where("status = ?", *status)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := db.Offset(offset).Limit(limit).Order("id desc").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}
