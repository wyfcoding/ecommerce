package mysql

import (
	"time"

	"github.com/wyfcoding/ecommerce/internal/subscription/domain"
	"gorm.io/gorm"
)

// SubscriptionPlanModel 订阅计划写模型。
type SubscriptionPlanModel struct {
	gorm.Model
	Name        string   `gorm:"column:name;type:varchar(128);not null;comment:计划名称"`
	Description string   `gorm:"column:description;type:varchar(255);comment:描述"`
	Price       uint64   `gorm:"column:price;not null;comment:价格(分)"`
	Duration    int32    `gorm:"column:duration;not null;comment:时长(天)"`
	Features    []string `gorm:"column:features;type:json;serializer:json;comment:特性列表"`
	Enabled     bool     `gorm:"column:enabled;not null;default:true;comment:是否启用"`
}

// SubscriptionModel 订阅记录写模型。
type SubscriptionModel struct {
	gorm.Model
	UserID     uint64                    `gorm:"column:user_id;not null;index;comment:用户ID"`
	PlanID     uint64                    `gorm:"column:plan_id;not null;index;comment:计划ID"`
	Status     domain.SubscriptionStatus `gorm:"column:status;type:tinyint;not null;default:1;comment:状态"`
	StartDate  time.Time                 `gorm:"column:start_date;not null;comment:开始时间"`
	EndDate    time.Time                 `gorm:"column:end_date;not null;comment:结束时间"`
	AutoRenew  bool                      `gorm:"column:auto_renew;not null;default:true;comment:自动续订"`
	CanceledAt *time.Time                `gorm:"column:canceled_at;comment:取消时间"`
}

func (SubscriptionPlanModel) TableName() string { return "subscription_plans" }
func (SubscriptionModel) TableName() string     { return "subscriptions" }

func toPlanModel(plan *domain.SubscriptionPlan) *SubscriptionPlanModel {
	if plan == nil {
		return nil
	}
	return &SubscriptionPlanModel{
		Model: gorm.Model{
			ID:        plan.ID,
			CreatedAt: plan.CreatedAt,
			UpdatedAt: plan.UpdatedAt,
		},
		Name:        plan.Name,
		Description: plan.Description,
		Price:       plan.Price,
		Duration:    plan.Duration,
		Features:    plan.Features,
		Enabled:     plan.Enabled,
	}
}

func toPlan(model *SubscriptionPlanModel) *domain.SubscriptionPlan {
	if model == nil {
		return nil
	}
	return &domain.SubscriptionPlan{
		ID:          model.ID,
		CreatedAt:   model.CreatedAt,
		UpdatedAt:   model.UpdatedAt,
		Name:        model.Name,
		Description: model.Description,
		Price:       model.Price,
		Duration:    model.Duration,
		Features:    model.Features,
		Enabled:     model.Enabled,
	}
}

func toSubscriptionModel(sub *domain.Subscription) *SubscriptionModel {
	if sub == nil {
		return nil
	}
	return &SubscriptionModel{
		Model: gorm.Model{
			ID:        sub.ID,
			CreatedAt: sub.CreatedAt,
			UpdatedAt: sub.UpdatedAt,
		},
		UserID:     sub.UserID,
		PlanID:     sub.PlanID,
		Status:     sub.Status,
		StartDate:  sub.StartDate,
		EndDate:    sub.EndDate,
		AutoRenew:  sub.AutoRenew,
		CanceledAt: sub.CanceledAt,
	}
}

func toSubscription(model *SubscriptionModel) *domain.Subscription {
	if model == nil {
		return nil
	}
	return &domain.Subscription{
		ID:         model.ID,
		CreatedAt:  model.CreatedAt,
		UpdatedAt:  model.UpdatedAt,
		UserID:     model.UserID,
		PlanID:     model.PlanID,
		Status:     model.Status,
		StartDate:  model.StartDate,
		EndDate:    model.EndDate,
		AutoRenew:  model.AutoRenew,
		CanceledAt: model.CanceledAt,
	}
}
