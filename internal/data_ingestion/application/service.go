package application

import (
	"context"
	"ecommerce/internal/data_ingestion/domain/entity"
	"ecommerce/internal/data_ingestion/domain/repository"
	"time"

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
		s.logger.Error("failed to save source", "error", err)
		return nil, err
	}
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
		return nil, err
	}

	// Async processing simulation
	go s.processJob(job)

	return job, nil
}

func (s *DataIngestionService) processJob(job *entity.IngestionJob) {
	ctx := context.Background()
	job.Start()
	s.repo.UpdateJob(ctx, job)

	// Simulate processing
	time.Sleep(2 * time.Second)

	// Success simulation
	job.Complete(100)
	s.repo.UpdateJob(ctx, job)

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
