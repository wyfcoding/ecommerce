package model

import "time"

// Event represents a generic event in the business logic layer.
type Event struct {
	EventType  string
	UserID     string
	EntityID   string
	Properties map[string]string
	Timestamp  time.Time
}
