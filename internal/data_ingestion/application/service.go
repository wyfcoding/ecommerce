package application

import (
	"context"
	"time"

	"github.com/wyfcoding/ecommerce/internal/data_ingestion/domain/entity"
	"github.com/wyfcoding/ecommerce/internal/data_ingestion/domain/repository"

	"log/slog"
)

type DataIngestionService struct {
	repo   repository.DataIngestionRepository
	logger *slog.Logger
}

func NewDataIngestionService(repo repository.DataIngestionRepository, logger *slog.Logger) *DataIngestionService {
	return &DataIngestionService{
		repo:   repo,
		logger: logger,
	}
}

// RegisterSource 注册数据源
func (s *DataIngestionService) RegisterSource(ctx context.Context, name string, sourceType entity.SourceType, config, description string) (*entity.IngestionSource, error) {
	source := entity.NewIngestionSource(name, sourceType, config, description)
	if err := s.repo.SaveSource(ctx, source); err != nil {
		s.logger.ErrorContext(ctx, "failed to save source", "name", name, "error", err)
		return nil, err
	}
	s.logger.InfoContext(ctx, "source registered successfully", "source_id", source.ID, "name", name)
	return source, nil
}

// ListSources 获取数据源列表
func (s *DataIngestionService) ListSources(ctx context.Context, page, pageSize int) ([]*entity.IngestionSource, int64, error) {
	offset := (page - 1) * pageSize
	return s.repo.ListSources(ctx, offset, pageSize)
}

// TriggerIngestion 触发数据摄取
func (s *DataIngestionService) TriggerIngestion(ctx context.Context, sourceID uint64) (*entity.IngestionJob, error) {
	source, err := s.repo.GetSource(ctx, sourceID)
	if err != nil {
		return nil, err
	}

	job := entity.NewIngestionJob(uint64(source.ID))
	if err := s.repo.SaveJob(ctx, job); err != nil {
		s.logger.ErrorContext(ctx, "failed to save job", "source_id", sourceID, "error", err)
		return nil, err
	}
	s.logger.InfoContext(ctx, "ingestion job triggered", "job_id", job.ID, "source_id", sourceID)

	// Async processing simulation
	go s.processJob(job)

	return job, nil
}

func (s *DataIngestionService) processJob(job *entity.IngestionJob) {
	ctx := context.Background()
	job.Start()
	s.repo.UpdateJob(ctx, job)
	s.logger.InfoContext(ctx, "ingestion job started", "job_id", job.ID)

	// Simulate processing
	time.Sleep(2 * time.Second)

	// Success simulation
	job.Complete(100)
	s.repo.UpdateJob(ctx, job)
	s.logger.InfoContext(ctx, "ingestion job completed", "job_id", job.ID, "rows_processed", 100)

	// Update source last run
	source, _ := s.repo.GetSource(ctx, job.SourceID)
	if source != nil {
		now := time.Now()
		source.LastRunAt = &now
		s.repo.UpdateSource(ctx, source)
	}
}

// ListJobs 获取任务列表
func (s *DataIngestionService) ListJobs(ctx context.Context, sourceID uint64, page, pageSize int) ([]*entity.IngestionJob, int64, error) {
	offset := (page - 1) * pageSize
	return s.repo.ListJobs(ctx, sourceID, offset, pageSize)
}
