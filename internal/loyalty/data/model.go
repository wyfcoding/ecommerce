package data

import (
	"time"

	"gorm.io/gorm"
)

// UserLoyaltyProfile is the database model for a user's loyalty profile.
type UserLoyaltyProfile struct {
	gorm.Model
	UserID          string    `gorm:"type:varchar(100);uniqueIndex;not null"`
	CurrentPoints   int64     `gorm:"default:0;not null"`
	LoyaltyLevel    string    `gorm:"type:varchar(50);default:'Bronze';not null"`
	LastLevelUpdate time.Time `gorm:"not null"`
}

// TableName specifies the table name for the UserLoyaltyProfile model.
func (UserLoyaltyProfile) TableName() string {
	return "user_loyalty_profiles"
}

// PointsTransaction is the database model for a points transaction.
type PointsTransaction struct {
	gorm.Model
	UserID       string    `gorm:"type:varchar(100);not null;index"`
	PointsChange int64     `gorm:"not null"`
	Reason       string    `gorm:"type:varchar(255);not null"`
	OrderID      string    `gorm:"type:varchar(100);index"` // Optional: link to an order
}

// TableName specifies the table name for the PointsTransaction model.
func (PointsTransaction) TableName() string {
	return "points_transactions"
}
