package repository

import (
	"context"
	"time"

	"ecommerce/internal/data_processing/model"
)

// DataProcessingRepo defines the interface for data processing data access.
type DataProcessingRepo interface {
	TriggerProcessingJob(ctx context.Context, jobType string, parameters map[string]string) (*model.ProcessingJob, error)
	TriggerSparkFlinkJob(ctx context.Context, jobName string, jobParameters map[string]string, platform string) (*model.ProcessingJob, error)
}