package biz

import (
	"context"
	"time"
)

// DataLakeRecord represents a record to be ingested into the data lake (Hudi) in the business logic layer.
type DataLakeRecord struct {
	TableName string
	RecordID  string
	Data      map[string]string
	EventTime time.Time
}

// DataLakeIngestionRepo defines the interface for data lake ingestion data access.
type DataLakeIngestionRepo interface {
	IngestData(ctx context.Context, tableName, recordID string, data map[string]string, eventTime time.Time) error
}

// DataLakeIngestionUsecase is the business logic for data lake ingestion.
type DataLakeIngestionUsecase struct {
	repo DataLakeIngestionRepo
}

// NewDataLakeIngestionUsecase creates a new DataLakeIngestionUsecase.
func NewDataLakeIngestionUsecase(repo DataLakeIngestionRepo) *DataLakeIngestionUsecase {
	return &DataLakeIngestionUsecase{repo: repo}
}

// IngestData ingests data into the data lake.
func (uc *DataLakeIngestionUsecase) IngestData(ctx context.Context, tableName, recordID string, data map[string]string, eventTime time.Time) error {
	// Add any business logic here, e.g., data validation, transformation
	return uc.repo.IngestData(ctx, tableName, recordID, data, eventTime)
}
