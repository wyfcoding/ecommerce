package service

import (
	"context"
	"time"

	"ecommerce/internal/data_processing/model"
	"ecommerce/internal/data_processing/repository"
)

// DataProcessingService is the business logic for data processing.
type DataProcessingService struct {
	repo repository.DataProcessingRepo
}

// NewDataProcessingService creates a new DataProcessingService.
func NewDataProcessingService(repo repository.DataProcessingRepo) *DataProcessingService {
	return &DataProcessingService{repo: repo}
}

// TriggerProcessingJob triggers a data processing job.
func (s *DataProcessingService) TriggerProcessingJob(ctx context.Context, jobType string, parameters map[string]string) (*model.ProcessingJob, error) {
	// Add any business logic here, e.g., validation, logging
	return s.repo.TriggerProcessingJob(ctx, jobType, parameters)
}

// TriggerSparkFlinkJob triggers a Spark/Flink job.
func (s *DataProcessingService) TriggerSparkFlinkJob(ctx context.Context, jobName string, jobParameters map[string]string, platform string) (*model.ProcessingJob, error) {
	// Add any business logic here, e.g., validation, logging
	return s.repo.TriggerSparkFlinkJob(ctx, jobName, jobParameters, platform)
}
