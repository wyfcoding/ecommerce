package repository

import (
	"time"

	"gorm.io/gorm"
)

// WishlistItem is the database model for an item in a user's wishlist.
type WishlistItem struct {
	gorm.Model
	UserID    string    `gorm:"type:varchar(100);not null;index"`
	ProductID string    `gorm:"type:varchar(100);not null;index"`
	AddedAt   time.Time `gorm:"not null"`
}

// TableName specifies the table name for the WishlistItem model.
func (WishlistItem) TableName() string {
	return "wishlist_items"
}
