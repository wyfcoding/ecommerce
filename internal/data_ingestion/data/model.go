package data

import (
	"time"
)

// Event represents a generic event for data ingestion.
type Event struct {
	EventType  string            `json:"event_type"`
	UserID     string            `json:"user_id"`
	EntityID   string            `json:"entity_id"`
	Properties map[string]string `json:"properties"`
	Timestamp  time.Time         `json:"timestamp"`
	IngestedAt time.Time         `json:"ingested_at"` // When the event was ingested
}
