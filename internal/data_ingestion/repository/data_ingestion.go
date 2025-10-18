package repository

import (
	"context"

	"ecommerce/internal/data_ingestion/model"
)

// DataIngestionRepo defines the interface for data ingestion data access.
type DataIngestionRepo interface {
	IngestEvent(ctx context.Context, event *model.Event) error
}