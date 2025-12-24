package application

import (
	"context"
	"log/slog"
	"time"

	"github.com/wyfcoding/ecommerce/internal/dataingestion/domain"
)

// DataIngestionManager 处理数据采集的写操作。
type DataIngestionManager struct {
	repo   domain.DataIngestionRepository
	logger *slog.Logger
}

// NewDataIngestionManager 创建并返回一个新的 DataIngestionManager 实例。
func NewDataIngestionManager(repo domain.DataIngestionRepository, logger *slog.Logger) *DataIngestionManager {
	return &DataIngestionManager{
		repo:   repo,
		logger: logger,
	}
}

// RegisterSource 注册一个新的数据源。
func (m *DataIngestionManager) RegisterSource(ctx context.Context, name string, sourceType domain.SourceType, config, description string) (*domain.IngestionSource, error) {
	source := domain.NewIngestionSource(name, sourceType, config, description)
	if err := m.repo.SaveSource(ctx, source); err != nil {
		m.logger.ErrorContext(ctx, "failed to save source", "name", name, "error", err)
		return nil, err
	}
	m.logger.InfoContext(ctx, "source registered successfully", "source_id", source.ID, "name", name)
	return source, nil
}

// TriggerIngestion 触发一个数据摄取任务。
func (m *DataIngestionManager) TriggerIngestion(ctx context.Context, sourceID uint64) (*domain.IngestionJob, error) {
	source, err := m.repo.GetSource(ctx, sourceID)
	if err != nil {
		return nil, err
	}

	job := domain.NewIngestionJob(uint64(source.ID))
	if err := m.repo.SaveJob(ctx, job); err != nil {
		m.logger.ErrorContext(ctx, "failed to save job", "source_id", sourceID, "error", err)
		return nil, err
	}
	m.logger.InfoContext(ctx, "ingestion job triggered", "job_id", job.ID, "source_id", sourceID)

	go m.processJob(job)

	return job, nil
}

// processJob 异步处理数据摄取任务的后台逻辑。
func (m *DataIngestionManager) processJob(job *domain.IngestionJob) {
	ctx := context.Background()
	job.Start()
	if err := m.repo.UpdateJob(ctx, job); err != nil {
		m.logger.ErrorContext(ctx, "failed to update job status to started", "job_id", job.ID, "error", err)
		return
	}
	m.logger.InfoContext(ctx, "ingestion job started", "job_id", job.ID)

	time.Sleep(2 * time.Second)

	job.Complete(100)
	if err := m.repo.UpdateJob(ctx, job); err != nil {
		m.logger.ErrorContext(ctx, "failed to update job status to completed", "job_id", job.ID, "error", err)
		return
	}
	m.logger.InfoContext(ctx, "ingestion job completed", "job_id", job.ID, "rows_processed", 100)

	source, _ := m.repo.GetSource(ctx, job.SourceID)
	if source != nil {
		now := time.Now()
		source.LastRunAt = &now
		if err := m.repo.UpdateSource(ctx, source); err != nil {
			m.logger.ErrorContext(ctx, "failed to update source last run time", "source_id", source.ID, "error", err)
		}
	}
}

// IngestEvent 摄取单个事件。
func (m *DataIngestionManager) IngestEvent(ctx context.Context, eventType, eventData, source string, timestamp time.Time) error {
	event := domain.NewIngestedEvent(eventType, eventData, source, timestamp)
	if err := m.repo.SaveEvent(ctx, event); err != nil {
		m.logger.ErrorContext(ctx, "failed to save event", "event_type", eventType, "error", err)
		return err
	}
	m.logger.InfoContext(ctx, "event ingested successfully", "event_id", event.ID, "event_type", eventType)
	return nil
}

// IngestBatch 批量摄取事件。
func (m *DataIngestionManager) IngestBatch(ctx context.Context, events []*domain.IngestedEvent) error {
	if len(events) == 0 {
		return nil
	}
	if err := m.repo.SaveEvents(ctx, events); err != nil {
		m.logger.ErrorContext(ctx, "failed to save batch events", "count", len(events), "error", err)
		return err
	}
	m.logger.InfoContext(ctx, "batch events ingested successfully", "count", len(events))
	return nil
}
