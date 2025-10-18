package model

import "time"

// UserProfile represents a simplified user profile for updates.
type UserProfile struct {
	UserID           string
	LastActiveTime   time.Time
	RecentCategories []string
	RecentBrands     []string
	// Add other fields that can be updated by real-time behavior
}
