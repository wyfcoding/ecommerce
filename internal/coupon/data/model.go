package data

import (
	"time"

	"gorm.io/gorm"
)

// Coupon is the database model for a coupon.
type Coupon struct {
	gorm.Model
	Code         string    `gorm:"type:varchar(100);uniqueIndex;not null"`
	Description  string    `gorm:"type:varchar(255)"`
	DiscountValue float64   `gorm:"not null"`
	DiscountType string    `gorm:"type:varchar(50);not null"` // e.g., FIXED_AMOUNT, PERCENTAGE
	ValidFrom    time.Time `gorm:"not null"`
	ValidTo      time.Time `gorm:"not null"`
	IsActive     bool      `gorm:"default:true"`
	MaxUsage     int       `gorm:"default:0"` // 0 for unlimited
	UsedCount    int       `gorm:"default:0"`
}

// TableName specifies the table name for the Coupon model.
func (Coupon) TableName() string {
	return "coupons"
}
