package data

import (
	"context"
	"ecommerce/internal/subscription/biz"

	"gorm.io/gorm"
)

// subscriptionRepo is the data layer implementation for SubscriptionRepo.
type subscriptionRepo struct {
	data *Data
	// log  *log.Helper
}

// toBiz converts a data.SubscriptionPlan model to a biz.SubscriptionPlan entity.
func (p *SubscriptionPlan) toBiz() *biz.SubscriptionPlan {
	if p == nil {
		return nil
	}
	return &biz.SubscriptionPlan{
		ID:             p.ID,
		Name:           p.Name,
		Description:    p.Description,
		Price:          p.Price,
		Currency:       p.Currency,
		RecurrenceType: p.RecurrenceType,
		DurationMonths: p.DurationMonths,
		IsActive:       p.IsActive,
		CreatedAt:      p.CreatedAt,
		UpdatedAt:      p.UpdatedAt,
	}
}

// fromBiz converts a biz.SubscriptionPlan entity to a data.SubscriptionPlan model.
func fromBizSubscriptionPlan(b *biz.SubscriptionPlan) *SubscriptionPlan {
	if b == nil {
		return nil
	}
	return &SubscriptionPlan{
		Name:           b.Name,
		Description:    b.Description,
		Price:          b.Price,
		Currency:       b.Currency,
		RecurrenceType: b.RecurrenceType,
		DurationMonths: b.DurationMonths,
		IsActive:       b.IsActive,
	}
}

// toBiz converts a data.UserSubscription model to a biz.UserSubscription entity.
func (s *UserSubscription) toBiz() *biz.UserSubscription {
	if s == nil {
		return nil
	}
	return &biz.UserSubscription{
		ID:              s.ID,
		UserID:          s.UserID,
		PlanID:          s.PlanID,
		Status:          s.Status,
		StartDate:       s.StartDate,
		EndDate:         s.EndDate,
		NextBillingDate: s.NextBillingDate,
		PaymentMethodID: s.PaymentMethodID,
		AutoRenew:       s.AutoRenew,
		CreatedAt:       s.CreatedAt,
		UpdatedAt:       s.UpdatedAt,
	}
}

// fromBiz converts a biz.UserSubscription entity to a data.UserSubscription model.
func fromBizUserSubscription(b *biz.UserSubscription) *UserSubscription {
	if b == nil {
		return nil
	}
	return &UserSubscription{
		UserID:          b.UserID,
		PlanID:          b.PlanID,
		Status:          b.Status,
		StartDate:       b.StartDate,
		EndDate:         b.EndDate,
		NextBillingDate: b.NextBillingDate,
		PaymentMethodID: b.PaymentMethodID,
		AutoRenew:       b.AutoRenew,
	}
}

// CreateSubscriptionPlan creates a new subscription plan in the database.
func (r *subscriptionRepo) CreateSubscriptionPlan(ctx context.Context, b *biz.SubscriptionPlan) (*biz.SubscriptionPlan, error) {
	plan := fromBizSubscriptionPlan(b)
	if err := r.data.db.WithContext(ctx).Create(plan).Error; err != nil {
		return nil, err
	}
	return plan.toBiz(), nil
}

// GetSubscriptionPlan retrieves a subscription plan by ID from the database.
func (r *subscriptionRepo) GetSubscriptionPlan(ctx context.Context, id uint) (*biz.SubscriptionPlan, error) {
	var plan SubscriptionPlan
	if err := r.data.db.WithContext(ctx).First(&plan, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, biz.ErrSubscriptionPlanNotFound
		}
		return nil, err
	}
	return plan.toBiz(), nil
}

// ListSubscriptionPlans lists subscription plans from the database with optional filters.
func (r *subscriptionRepo) ListSubscriptionPlans(ctx context.Context, activeOnly bool) ([]*biz.SubscriptionPlan, int32, error) {
	var plans []SubscriptionPlan
	var totalCount int32

	query := r.data.db.WithContext(ctx).Model(&SubscriptionPlan{})

	if activeOnly {
		query = query.Where("is_active = ?", true)
	}

	// Get total count
	query.Count(int64(&totalCount))

	if err := query.Find(&plans).Error; err != nil {
		return nil, 0, err
	}

	bizPlans := make([]*biz.SubscriptionPlan, len(plans))
	for i, p := range plans {
		bizPlans[i] = p.toBiz()
	}

	return bizPlans, totalCount, nil
}

// UpdateSubscriptionPlan updates an existing subscription plan in the database.
func (r *subscriptionRepo) UpdateSubscriptionPlan(ctx context.Context, b *biz.SubscriptionPlan) (*biz.SubscriptionPlan, error) {
	plan := fromBizSubscriptionPlan(b)
	plan.ID = b.ID // Ensure ID is set for update
	if err := r.data.db.WithContext(ctx).Save(plan).Error; err != nil {
		return nil, err
	}
	return plan.toBiz(), nil
}

// CreateUserSubscription creates a new user subscription in the database.
func (r *subscriptionRepo) CreateUserSubscription(ctx context.Context, b *biz.UserSubscription) (*biz.UserSubscription, error) {
	sub := fromBizUserSubscription(b)
	if err := r.data.db.WithContext(ctx).Create(sub).Error; err != nil {
		return nil, err
	}
	return sub.toBiz(), nil
}

// GetUserSubscription retrieves a user subscription by ID from the database.
func (r *subscriptionRepo) GetUserSubscription(ctx context.Context, id uint) (*biz.UserSubscription, error) {
	var sub UserSubscription
	if err := r.data.db.WithContext(ctx).First(&sub, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, biz.ErrUserSubscriptionNotFound
		}
		return nil, err
	}
	return sub.toBiz(), nil
}

// GetUserActiveSubscriptionByUserID retrieves a user's active subscription by user ID from the database.
func (r *subscriptionRepo) GetUserActiveSubscriptionByUserID(ctx context.Context, userID string) (*biz.UserSubscription, error) {
	var sub UserSubscription
	if err := r.data.db.WithContext(ctx).Where("user_id = ? AND status = ?", userID, "ACTIVE").First(&sub).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, biz.ErrUserSubscriptionNotFound
		}
		return nil, err
	}
	return sub.toBiz(), nil
}

// UpdateUserSubscription updates an existing user subscription in the database.
func (r *subscriptionRepo) UpdateUserSubscription(ctx context.Context, b *biz.UserSubscription) (*biz.UserSubscription, error) {
	sub := fromBizUserSubscription(b)
	sub.ID = b.ID // Ensure ID is set for update
	if err := r.data.db.WithContext(ctx).Save(sub).Error; err != nil {
		return nil, err
	}
	return sub.toBiz(), nil
}
