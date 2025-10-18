package model

import "time"

// SubscriptionPlan represents a subscription plan in the business layer.
type SubscriptionPlan struct {
	ID             uint
	Name           string
	Description    string
	Price          float64
	Currency       string
	RecurrenceType string // e.g., MONTHLY, ANNUALLY
	DurationMonths int32  // 0 for indefinite
	IsActive       bool
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// UserSubscription represents a user's subscription to a plan in the business layer.
type UserSubscription struct {
	ID              uint
	UserID          string
	PlanID          string
	Status          string // e.g., ACTIVE, CANCELED, EXPIRED, PENDING_RENEWAL
	StartDate       time.Time
	EndDate         time.Time
	NextBillingDate time.Time
	PaymentMethodID string
	AutoRenew       bool
	CreatedAt       time.Time
	UpdatedAt       time.Time
}