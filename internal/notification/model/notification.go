package model

import "time"

// Notification represents a notification in the business logic layer.
type Notification struct {
	ID             uint
	NotificationID string
	UserID         uint64
	Type           string
	Title          string
	Content        string
	IsRead         bool
	CreatedAt      time.Time
}
