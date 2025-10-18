package model

import "time"

// UserLoyaltyProfile represents a user's loyalty profile in the business layer.
type UserLoyaltyProfile struct {
	UserID          string
	CurrentPoints   int64
	LoyaltyLevel    string
	LastLevelUpdate time.Time
}

// PointsTransaction represents a points transaction in the business layer.
type PointsTransaction struct {
	ID           uint
	UserID       string
	PointsChange int64
	Reason       string
	OrderID      string
	CreatedAt    time.Time
}