package biz

import (
	"context"
	"time"
)

// Event represents a generic event in the business logic layer.
type Event struct {
	EventType  string
	UserID     string
	EntityID   string
	Properties map[string]string
	Timestamp  time.Time
}

// DataIngestionRepo defines the interface for data ingestion data access.
type DataIngestionRepo interface {
	IngestEvent(ctx context.Context, event *Event) error
}

// DataIngestionUsecase is the business logic for data ingestion.
type DataIngestionUsecase struct {
	repo DataIngestionRepo
}

// NewDataIngestionUsecase creates a new DataIngestionUsecase.
func NewDataIngestionUsecase(repo DataIngestionRepo) *DataIngestionUsecase {
	return &DataIngestionUsecase{repo: repo}
}

// IngestEvent ingests a generic event.
func (uc *DataIngestionUsecase) IngestEvent(ctx context.Context, eventType, userID, entityID string, properties map[string]string, timestamp time.Time) error {
	event := &Event{
		EventType:  eventType,
		UserID:     userID,
		EntityID:   entityID,
		Properties: properties,
		Timestamp:  timestamp,
	}
	return uc.repo.IngestEvent(ctx, event)
}
