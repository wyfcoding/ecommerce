package persistence

import (
	"context"
	"ecommerce/internal/subscription/domain/entity"
	"ecommerce/internal/subscription/domain/repository"
	"errors"
	"time"

	"gorm.io/gorm"
)

type subscriptionRepository struct {
	db *gorm.DB
}

func NewSubscriptionRepository(db *gorm.DB) repository.SubscriptionRepository {
	return &subscriptionRepository{db: db}
}

// 订阅计划
func (r *subscriptionRepository) SavePlan(ctx context.Context, plan *entity.SubscriptionPlan) error {
	return r.db.WithContext(ctx).Save(plan).Error
}

func (r *subscriptionRepository) GetPlan(ctx context.Context, id uint64) (*entity.SubscriptionPlan, error) {
	var plan entity.SubscriptionPlan
	if err := r.db.WithContext(ctx).First(&plan, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &plan, nil
}

func (r *subscriptionRepository) ListPlans(ctx context.Context, enabledOnly bool) ([]*entity.SubscriptionPlan, error) {
	var list []*entity.SubscriptionPlan
	db := r.db.WithContext(ctx)
	if enabledOnly {
		db = db.Where("enabled = ?", true)
	}
	if err := db.Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

// 订阅
func (r *subscriptionRepository) SaveSubscription(ctx context.Context, sub *entity.Subscription) error {
	return r.db.WithContext(ctx).Save(sub).Error
}

func (r *subscriptionRepository) GetSubscription(ctx context.Context, id uint64) (*entity.Subscription, error) {
	var sub entity.Subscription
	if err := r.db.WithContext(ctx).First(&sub, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &sub, nil
}

func (r *subscriptionRepository) GetActiveSubscription(ctx context.Context, userID uint64) (*entity.Subscription, error) {
	var sub entity.Subscription
	now := time.Now()
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND status = ? AND end_date > ?", userID, entity.SubscriptionStatusActive, now).
		First(&sub).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &sub, nil
}

func (r *subscriptionRepository) ListSubscriptions(ctx context.Context, userID uint64, status *entity.SubscriptionStatus, offset, limit int) ([]*entity.Subscription, int64, error) {
	var list []*entity.Subscription
	var total int64

	db := r.db.WithContext(ctx).Model(&entity.Subscription{})
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
