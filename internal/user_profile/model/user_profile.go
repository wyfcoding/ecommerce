package model

import "time"

// UserProfile represents a user's profile in the business logic layer.
type UserProfile struct {
	UserID           string
	Gender           string
	AgeGroup         string
	City             string
	Interests        []string
	RecentCategories []string
	RecentBrands     []string
	TotalSpent       uint64
	OrderCount       uint32
	LastActiveTime   time.Time
	CustomTags       map[string]string
	UpdatedAt        time.Time
}

// UserBehaviorEvent represents a user behavior event in the business logic layer.
type UserBehaviorEvent struct {
	UserID       string
	BehaviorType string
	ItemID       string
	Properties   map[string]string
	EventTime    time.Time
}
