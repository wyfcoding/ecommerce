package mysql

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

// --- tx helpers ---

func (r *subscriptionRepository) BeginTx(ctx context.Context) any {
	return r.db.WithContext(ctx).Begin()
}

func (r *subscriptionRepository) CommitTx(tx any) error {
	gormTx, ok := tx.(*gorm.DB)
	if !ok || gormTx == nil {
		return errors.New("invalid transaction")
	}
	return gormTx.Commit().Error
}

func (r *subscriptionRepository) RollbackTx(tx any) error {
	gormTx, ok := tx.(*gorm.DB)
	if !ok || gormTx == nil {
		return errors.New("invalid transaction")
	}
	return gormTx.Rollback().Error
}

func (r *subscriptionRepository) WithTx(ctx context.Context, fn func(tx any) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(tx)
	})
}

// --- SubscriptionPlan methods ---

func (r *subscriptionRepository) SavePlan(ctx context.Context, plan *domain.SubscriptionPlan) error {
	return r.savePlanWithTx(ctx, r.db, plan)
}

func (r *subscriptionRepository) SavePlanInTx(ctx context.Context, tx any, plan *domain.SubscriptionPlan) error {
	gormTx, ok := tx.(*gorm.DB)
	if !ok || gormTx == nil {
		return errors.New("invalid transaction")
	}
	return r.savePlanWithTx(ctx, gormTx, plan)
}

func (r *subscriptionRepository) GetPlan(ctx context.Context, id uint64) (*domain.SubscriptionPlan, error) {
	var plan SubscriptionPlanModel
	if err := r.db.WithContext(ctx).First(&plan, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toPlan(&plan), nil
}

func (r *subscriptionRepository) ListPlans(ctx context.Context, enabledOnly bool) ([]*domain.SubscriptionPlan, error) {
	var list []*SubscriptionPlanModel
	db := r.db.WithContext(ctx).Model(&SubscriptionPlanModel{})
	if enabledOnly {
		db = db.Where("enabled = ?", true)
	}
	if err := db.Order("created_at desc").Find(&list).Error; err != nil {
		return nil, err
	}
	items := make([]*domain.SubscriptionPlan, len(list))
	for i, p := range list {
		items[i] = toPlan(p)
	}
	return items, nil
}

func (r *subscriptionRepository) DeletePlan(ctx context.Context, id uint64) error {
	return r.deletePlanWithTx(ctx, r.db, id)
}

func (r *subscriptionRepository) DeletePlanInTx(ctx context.Context, tx any, id uint64) error {
	gormTx, ok := tx.(*gorm.DB)
	if !ok || gormTx == nil {
		return errors.New("invalid transaction")
	}
	return r.deletePlanWithTx(ctx, gormTx, id)
}

// --- Subscription methods ---

func (r *subscriptionRepository) SaveSubscription(ctx context.Context, sub *domain.Subscription) error {
	return r.saveSubscriptionWithTx(ctx, r.db, sub)
}

func (r *subscriptionRepository) SaveSubscriptionInTx(ctx context.Context, tx any, sub *domain.Subscription) error {
	gormTx, ok := tx.(*gorm.DB)
	if !ok || gormTx == nil {
		return errors.New("invalid transaction")
	}
	return r.saveSubscriptionWithTx(ctx, gormTx, sub)
}

func (r *subscriptionRepository) GetSubscription(ctx context.Context, id uint64) (*domain.Subscription, error) {
	var sub SubscriptionModel
	if err := r.db.WithContext(ctx).First(&sub, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toSubscription(&sub), nil
}

func (r *subscriptionRepository) GetActiveSubscription(ctx context.Context, userID uint64) (*domain.Subscription, error) {
	var sub SubscriptionModel
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
	return toSubscription(&sub), nil
}

func (r *subscriptionRepository) ListSubscriptions(ctx context.Context, query *domain.SubscriptionQuery) ([]*domain.Subscription, int64, error) {
	var list []*SubscriptionModel
	var total int64

	db := r.db.WithContext(ctx).Model(&SubscriptionModel{})
	if query != nil {
		if query.UserID > 0 {
			db = db.Where("user_id = ?", query.UserID)
		}
		if query.PlanID > 0 {
			db = db.Where("plan_id = ?", query.PlanID)
		}
		if query.Status != nil {
			db = db.Where("status = ?", *query.Status)
		}
		if query.AutoRenew != nil {
			db = db.Where("auto_renew = ?", *query.AutoRenew)
		}
		if !query.StartTime.IsZero() {
			db = db.Where("created_at >= ?", query.StartTime)
		}
		if !query.EndTime.IsZero() {
			db = db.Where("created_at <= ?", query.EndTime)
		}
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	page := 1
	pageSize := 20
	if query != nil {
		if query.Page > 0 {
			page = query.Page
		}
		if query.PageSize > 0 {
			pageSize = query.PageSize
		}
	}
	offset := (page - 1) * pageSize
	if err := db.Offset(offset).Limit(pageSize).Order("created_at desc").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	items := make([]*domain.Subscription, len(list))
	for i, sub := range list {
		items[i] = toSubscription(sub)
	}
	return items, total, nil
}

func (r *subscriptionRepository) DeleteSubscription(ctx context.Context, id uint64) error {
	return r.deleteSubscriptionWithTx(ctx, r.db, id)
}

func (r *subscriptionRepository) DeleteSubscriptionInTx(ctx context.Context, tx any, id uint64) error {
	gormTx, ok := tx.(*gorm.DB)
	if !ok || gormTx == nil {
		return errors.New("invalid transaction")
	}
	return r.deleteSubscriptionWithTx(ctx, gormTx, id)
}

// --- internal helpers ---

func (r *subscriptionRepository) savePlanWithTx(ctx context.Context, tx *gorm.DB, plan *domain.SubscriptionPlan) error {
	if plan == nil {
		return nil
	}
	model := toPlanModel(plan)
	if err := tx.WithContext(ctx).Save(model).Error; err != nil {
		return err
	}
	if synced := toPlan(model); synced != nil {
		*plan = *synced
	}
	return nil
}

func (r *subscriptionRepository) deletePlanWithTx(ctx context.Context, tx *gorm.DB, id uint64) error {
	return tx.WithContext(ctx).Delete(&SubscriptionPlanModel{}, id).Error
}

func (r *subscriptionRepository) saveSubscriptionWithTx(ctx context.Context, tx *gorm.DB, sub *domain.Subscription) error {
	if sub == nil {
		return nil
	}
	model := toSubscriptionModel(sub)
	if err := tx.WithContext(ctx).Save(model).Error; err != nil {
		return err
	}
	if synced := toSubscription(model); synced != nil {
		*sub = *synced
	}
	return nil
}

func (r *subscriptionRepository) deleteSubscriptionWithTx(ctx context.Context, tx *gorm.DB, id uint64) error {
	return tx.WithContext(ctx).Delete(&SubscriptionModel{}, id).Error
}
