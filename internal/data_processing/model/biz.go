package biz

import (
	"context"
	"time"
)

// ProcessingJob represents a data processing job in the business logic layer.
type ProcessingJob struct {
	JobID       string
	JobType     string
	Parameters  map[string]string
	Status      string
	TriggeredAt time.Time
	CompletedAt *time.Time
}

// DataProcessingRepo defines the interface for data processing data access.
type DataProcessingRepo interface {
	TriggerProcessingJob(ctx context.Context, jobType string, parameters map[string]string) (*ProcessingJob, error)
	TriggerSparkFlinkJob(ctx context.Context, jobName string, jobParameters map[string]string, platform string) (*ProcessingJob, error)
}

// DataProcessingUsecase is the business logic for data processing.
type DataProcessingUsecase struct {
	repo DataProcessingRepo
}

// NewDataProcessingUsecase creates a new DataProcessingUsecase.
func NewDataProcessingUsecase(repo DataProcessingRepo) *DataProcessingUsecase {
	return &DataProcessingUsecase{repo: repo}
}

// TriggerProcessingJob triggers a data processing job.
func (uc *DataProcessingUsecase) TriggerProcessingJob(ctx context.Context, jobType string, parameters map[string]string) (*ProcessingJob, error) {
	// Add any business logic here, e.g., validation, logging
	return uc.repo.TriggerProcessingJob(ctx, jobType, parameters)
}

// TriggerSparkFlinkJob triggers a Spark/Flink job.
func (uc *DataProcessingUsecase) TriggerSparkFlinkJob(ctx context.Context, jobName string, jobParameters map[string]string, platform string) (*ProcessingJob, error) {
	// Add any business logic here, e.g., validation, logging
	return uc.repo.TriggerSparkFlinkJob(ctx, jobName, jobParameters, platform)
}
