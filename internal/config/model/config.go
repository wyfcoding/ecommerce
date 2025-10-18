package model

import "time"

// ConfigEntry represents a configuration entry in the business logic layer.
type ConfigEntry struct {
	ID          uint
	Key         string
	Value       string
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
