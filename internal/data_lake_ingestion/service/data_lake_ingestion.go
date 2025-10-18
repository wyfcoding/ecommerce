package service

import (
	"context"
	"time"

	"ecommerce/internal/data_lake_ingestion/model"
	"ecommerce/internal/data_lake_ingestion/repository"
)

// DataLakeIngestionService is the business logic for data lake ingestion.
type DataLakeIngestionService struct {
	repo repository.DataLakeIngestionRepo
}

// NewDataLakeIngestionService creates a new DataLakeIngestionService.
func NewDataLakeIngestionService(repo repository.DataLakeIngestionRepo) *DataLakeIngestionService {
	return &DataLakeIngestionService{repo: repo}
}

// IngestData ingests data into the data lake.
func (s *DataLakeIngestionService) IngestData(ctx context.Context, tableName, recordID string, data map[string]string, eventTime time.Time) error {
	// Add any business logic here, e.g., data validation, transformation
	return s.repo.IngestData(ctx, tableName, recordID, data, eventTime)
}
