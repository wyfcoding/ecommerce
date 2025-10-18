package model

import "time"

// WishlistItem represents an item in a user's wishlist in the business layer.
type WishlistItem struct {
	ID        uint
	UserID    string
	ProductID string
	AddedAt   time.Time
}