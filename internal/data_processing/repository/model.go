package data

import (
	"time"
)

// ProcessingJob represents a data processing job.
type ProcessingJob struct {
	JobID       string            `json:"job_id"`
	JobType     string            `json:"job_type"`
	Parameters  map[string]string `json:"parameters"`
	Status      string            `json:"status"` // e.g., "TRIGGERED", "RUNNING", "COMPLETED", "FAILED"
	TriggeredAt time.Time         `json:"triggered_at"`
	CompletedAt *time.Time        `json:"completed_at,omitempty"`
}
