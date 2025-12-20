package application

import (
	"context"
	"time"

	"github.com/wyfcoding/ecommerce/internal/data_ingestion/domain"
)

// DataIngestionService acts as a facade for data ingestion operations.
type DataIngestionService struct {
	manager *DataIngestionManager
	query   *DataIngestionQuery
}

// NewDataIngestionService creates a new DataIngestionService facade.
func NewDataIngestionService(manager *DataIngestionManager, query *DataIngestionQuery) *DataIngestionService {
	return &DataIngestionService{
		manager: manager,
		query:   query,
	}
}

// --- Write Operations (Delegated to Manager) ---

func (s *DataIngestionService) RegisterSource(ctx context.Context, name string, sourceType domain.SourceType, config, description string) (*domain.IngestionSource, error) {
	return s.manager.RegisterSource(ctx, name, sourceType, config, description)
}

func (s *DataIngestionService) TriggerIngestion(ctx context.Context, sourceID uint64) (*domain.IngestionJob, error) {
	return s.manager.TriggerIngestion(ctx, sourceID)
}

func (s *DataIngestionService) IngestEvent(ctx context.Context, eventType, eventData, source string, timestamp time.Time) error {
	return s.manager.IngestEvent(ctx, eventType, eventData, source, timestamp)
}

func (s *DataIngestionService) IngestBatch(ctx context.Context, events []*domain.IngestedEvent) error {
	return s.manager.IngestBatch(ctx, events)
}

// --- Read Operations (Delegated to Query) ---

func (s *DataIngestionService) ListSources(ctx context.Context, page, pageSize int) ([]*domain.IngestionSource, int64, error) {
	return s.query.ListSources(ctx, page, pageSize)
}

func (s *DataIngestionService) ListJobs(ctx context.Context, sourceID uint64, page, pageSize int) ([]*domain.IngestionJob, int64, error) {
	return s.query.ListJobs(ctx, sourceID, page, pageSize)
}
