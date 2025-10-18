package model

import "time"

// ProcessingJob represents a data processing job in the business logic layer.
type ProcessingJob struct {
	JobID       string
	JobType     string
	Parameters  map[string]string
	Status      string
	TriggeredAt time.Time
	CompletedAt *time.Time
}
