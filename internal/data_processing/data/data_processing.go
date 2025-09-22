package data

import (
	"context"
	"ecommerce/internal/data_processing/biz"
	"ecommerce/internal/data_processing/data/model"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type dataProcessingRepo struct {
	data *Data // Placeholder for common data dependencies if any
}

// NewDataProcessingRepo creates a new DataProcessingRepo.
func NewDataProcessingRepo(data *Data) biz.DataProcessingRepo {
	return &dataProcessingRepo{data: data}
}

// TriggerProcessingJob simulates triggering a data processing job.
func (r *dataProcessingRepo) TriggerProcessingJob(ctx context.Context, jobType string, parameters map[string]string) (*biz.ProcessingJob, error) {
	jobID := uuid.New().String()
	triggeredAt := time.Now()

	// Simulate job triggering and status update
	job := &model.ProcessingJob{
		JobID:       jobID,
		JobType:     jobType,
		Parameters:  parameters,
		Status:      "TRIGGERED", // Initial status
		TriggeredAt: triggeredAt,
	}

	// In a real system, this would interact with a job orchestration system
	// For now, just log the job details
	fmt.Printf("Data processing job triggered: JobID=%s, Type=%s, Parameters=%v\n", job.JobID, job.JobType, job.Parameters)

	return &biz.ProcessingJob{
		JobID:       job.JobID,
		JobType:     job.JobType,
		Parameters:  job.Parameters,
		Status:      job.Status,
		TriggeredAt: job.TriggeredAt,
	}, nil
}
