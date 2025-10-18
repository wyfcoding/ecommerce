package model

import "time"

// ModerationResult represents the result of a content moderation check in the business logic layer.
type ModerationResult struct {
	ID          uint
	ContentID   string
	ContentType string
	UserID      string
	TextContent string
	ImageURL    string
	IsSafe      bool
	Labels      []string
	Confidence  float64
	Decision    string
	ModeratedAt time.Time
}
