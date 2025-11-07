package repository

import (
	"time"

	"gorm.io/gorm"
)

// SubscriptionPlan is the database model for a subscription plan.
type SubscriptionPlan struct {
	gorm.Model
	Name           string  `gorm:"type:varchar(255);uniqueIndex;not null"`
	Description    string  `gorm:"type:text"`
	Price          float64 `gorm:"not null"`
	Currency       string  `gorm:"type:varchar(10);not null"`
	RecurrenceType string  `gorm:"type:varchar(50);not null"` // e.g., MONTHLY, ANNUALLY
	DurationMonths int32   `gorm:"default:0"`                 // 0 for indefinite
	IsActive       bool    `gorm:"default:true"`
}

// TableName specifies the table name for the SubscriptionPlan model.
func (SubscriptionPlan) TableName() string {
	return "subscription_plans"
}

// UserSubscription is the database model for a user's subscription to a plan.
type UserSubscription struct {
	gorm.Model
	UserID          string    `gorm:"type:varchar(100);not null;index"`
	PlanID          string    `gorm:"type:varchar(100);not null"` // Reference to SubscriptionPlan.ID
	Status          string    `gorm:"type:varchar(50);not null"`  // e.g., ACTIVE, CANCELED, EXPIRED, PENDING_RENEWAL
	StartDate       time.Time `gorm:"not null"`
	EndDate         time.Time `gorm:"not null"`
	NextBillingDate time.Time `gorm:"not null"`
	PaymentMethodID string    `gorm:"type:varchar(100)"`
	AutoRenew       bool      `gorm:"default:true"`
}

// TableName specifies the table name for the UserSubscription model.
func (UserSubscription) TableName() string {
	return "user_subscriptions"
}
