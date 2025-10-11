package data

import (
	"context"
	"ecommerce/internal/data_ingestion/biz"
	"ecommerce/internal/data_ingestion/data/model"
	"encoding/json"
	"fmt"
	"os"
	"time"
)

const (
	dataLakePath = "/tmp/data_lake_events.log" // Placeholder for HDFS/S3 path
)

type dataIngestionRepo struct {
	data *Data // Placeholder for common data dependencies if any
}

// NewDataIngestionRepo creates a new DataIngestionRepo.
func NewDataIngestionRepo(data *Data) biz.DataIngestionRepo {
	return &dataIngestionRepo{data: data}
}

// IngestEvent simulates ingesting an event into a data lake.
func (r *dataIngestionRepo) IngestEvent(ctx context.Context, event *biz.Event) error {
	// Convert biz.Event to data.model.Event
	dataEvent := model.Event{
		EventType:  event.EventType,
		UserID:     event.UserID,
		EntityID:   event.EntityID,
		Properties: event.Properties,
		Timestamp:  event.Timestamp,
		IngestedAt: time.Now(),
	}

	// Simulate writing to a data lake (e.g., HDFS, S3, or a file for simplicity)
	file, err := os.OpenFile(dataLakePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open data lake file: %w", err)
	}
	defer file.Close()

	eventBytes, err := json.Marshal(dataEvent)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	_, err = file.WriteString(string(eventBytes) + "\n")
	if err != nil {
		return fmt.Errorf("failed to write event to data lake file: %w", err)
	}

	fmt.Printf("Event ingested: %s\n", event.EventType) // For demonstration
	return nil
}
