package service

import (
	"context"
	"time"

	"ecommerce/internal/data_ingestion/model"
	"ecommerce/internal/data_ingestion/repository"
)

// DataIngestionService is the business logic for data ingestion.
type DataIngestionService struct {
	repo repository.DataIngestionRepo
}

// NewDataIngestionService creates a new DataIngestionService.
func NewDataIngestionService(repo repository.DataIngestionRepo) *DataIngestionService {
	return &DataIngestionService{repo: repo}
}

// IngestEvent ingests a generic event.
func (s *DataIngestionService) IngestEvent(ctx context.Context, eventType, userID, entityID string, properties map[string]string, timestamp time.Time) error {
	event := &model.Event{
		EventType:  eventType,
		UserID:     userID,
		EntityID:   entityID,
		Properties: properties,
		Timestamp:  timestamp,
	}
	return s.repo.IngestEvent(ctx, event)
}
