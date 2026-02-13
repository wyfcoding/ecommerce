package domain

import "time"

type DomainEvent interface {
	EventName() string
	OccurredAt() time.Time
}

type DataIngestedEvent struct {
	IngestionID uint64    `json:"ingestion_id"`
	Source      string    `json:"source"`
	DataType    string    `json:"data_type"`
	Timestamp   time.Time `json:"timestamp"`
}

func (e *DataIngestedEvent) EventName() string     { return "dataingestion.ingested" }
func (e *DataIngestedEvent) OccurredAt() time.Time { return e.Timestamp }

type IngestionFailedEvent struct {
	IngestionID uint64    `json:"ingestion_id"`
	Source      string    `json:"source"`
	Reason      string    `json:"reason"`
	Timestamp   time.Time `json:"timestamp"`
}

func (e *IngestionFailedEvent) EventName() string     { return "dataingestion.failed" }
func (e *IngestionFailedEvent) OccurredAt() time.Time { return e.Timestamp }

type SourceCreatedEvent struct {
	SourceID   uint64    `json:"source_id"`
	Name       string    `json:"name"`
	SourceType SourceType `json:"source_type"`
	Timestamp  time.Time `json:"timestamp"`
}

func (e *SourceCreatedEvent) EventName() string     { return "dataingestion.source.created" }
func (e *SourceCreatedEvent) OccurredAt() time.Time { return e.Timestamp }

type SourceActivatedEvent struct {
	SourceID  uint64    `json:"source_id"`
	Timestamp time.Time `json:"timestamp"`
}

func (e *SourceActivatedEvent) EventName() string     { return "dataingestion.source.activated" }
func (e *SourceActivatedEvent) OccurredAt() time.Time { return e.Timestamp }

type SourceDeactivatedEvent struct {
	SourceID  uint64    `json:"source_id"`
	Reason    string    `json:"reason"`
	Timestamp time.Time `json:"timestamp"`
}

func (e *SourceDeactivatedEvent) EventName() string     { return "dataingestion.source.deactivated" }
func (e *SourceDeactivatedEvent) OccurredAt() time.Time { return e.Timestamp }

type JobStartedEvent struct {
	JobID     uint64    `json:"job_id"`
	SourceID  uint64    `json:"source_id"`
	Timestamp time.Time `json:"timestamp"`
}

func (e *JobStartedEvent) EventName() string     { return "dataingestion.job.started" }
func (e *JobStartedEvent) OccurredAt() time.Time { return e.Timestamp }

type JobCompletedEvent struct {
	JobID            uint64    `json:"job_id"`
	SourceID         uint64    `json:"source_id"`
	RecordsProcessed int64     `json:"records_processed"`
	Duration         int64     `json:"duration_ms"`
	Timestamp        time.Time `json:"timestamp"`
}

func (e *JobCompletedEvent) EventName() string     { return "dataingestion.job.completed" }
func (e *JobCompletedEvent) OccurredAt() time.Time { return e.Timestamp }
