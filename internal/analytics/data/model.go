package data

import (
	"time"
)

// ProductViewEvent represents a product view event for ClickHouse.
type ProductViewEvent struct {
	UserID    uint64    `ch:"user_id"`
	ProductID uint64    `ch:"product_id"`
	ViewTime  time.Time `ch:"view_time"`
	// Add other relevant fields like IP address, user agent, etc.
}
