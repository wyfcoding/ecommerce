package repository

import (
	"context"
	"time"

	"ecommerce/internal/data_lake_ingestion/model"
)

// DataLakeIngestionRepo defines the interface for data lake ingestion data access.
type DataLakeIngestionRepo interface {
	IngestData(ctx context.Context, tableName, recordID string, data map[string]string, eventTime time.Time) error
}