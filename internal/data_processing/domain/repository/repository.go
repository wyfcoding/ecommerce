package repository

import (
	"context"
	"ecommerce/internal/data_processing/domain/entity"
)

// DataProcessingRepository 数据处理仓储接口
type DataProcessingRepository interface {
	// Task methods
	SaveTask(ctx context.Context, task *entity.ProcessingTask) error
	GetTask(ctx context.Context, id uint64) (*entity.ProcessingTask, error)
	ListTasks(ctx context.Context, workflowID uint64, status entity.TaskStatus, offset, limit int) ([]*entity.ProcessingTask, int64, error)
	UpdateTask(ctx context.Context, task *entity.ProcessingTask) error

	// Workflow methods
	SaveWorkflow(ctx context.Context, workflow *entity.ProcessingWorkflow) error
	GetWorkflow(ctx context.Context, id uint64) (*entity.ProcessingWorkflow, error)
	ListWorkflows(ctx context.Context, offset, limit int) ([]*entity.ProcessingWorkflow, int64, error)
	UpdateWorkflow(ctx context.Context, workflow *entity.ProcessingWorkflow) error
}
