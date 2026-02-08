package application

import (
	"context"
	"log/slog"

	"github.com/wyfcoding/ecommerce/internal/dataprocessing/domain"
	"github.com/wyfcoding/pkg/messagequeue"
)

// DataProcessingService 数据处理门面服务，整合 Command 和 Query。
type DataProcessingService struct {
	Command *DataProcessingCommandService
	Query   *DataProcessingQueryService
}

// NewDataProcessingService 构造函数。
func NewDataProcessingService(repo domain.DataProcessingRepository, publisher messagequeue.EventPublisher, logger *slog.Logger) *DataProcessingService {
	return &DataProcessingService{
		Command: NewDataProcessingCommandService(repo, publisher, logger),
		Query:   NewDataProcessingQueryService(repo),
	}
}

// --- Manager (Writes) ---

func (s *DataProcessingService) SubmitTask(ctx context.Context, name, taskType, config string, workflowID uint64) (*domain.ProcessingTask, error) {
	return s.Command.SubmitTask(ctx, name, taskType, config, workflowID)
}

func (s *DataProcessingService) CreateWorkflow(ctx context.Context, name, description, steps string) (*domain.ProcessingWorkflow, error) {
	return s.Command.CreateWorkflow(ctx, name, description, steps)
}

func (s *DataProcessingService) CancelTask(ctx context.Context, id uint64) error {
	return s.Command.CancelTask(ctx, id)
}

func (s *DataProcessingService) UpdateWorkflow(ctx context.Context, id uint64, name, description, steps string) error {
	return s.Command.UpdateWorkflow(ctx, id, name, description, steps)
}

func (s *DataProcessingService) SetWorkflowActive(ctx context.Context, id uint64, active bool) error {
	return s.Command.SetWorkflowActive(ctx, id, active)
}

// --- Query (Reads) ---

func (s *DataProcessingService) GetTask(ctx context.Context, id uint64) (*domain.ProcessingTask, error) {
	return s.Query.GetTask(ctx, id)
}

func (s *DataProcessingService) GetWorkflow(ctx context.Context, id uint64) (*domain.ProcessingWorkflow, error) {
	return s.Query.GetWorkflow(ctx, id)
}

func (s *DataProcessingService) ListTasks(ctx context.Context, workflowID uint64, status domain.TaskStatus, page, pageSize int) ([]*domain.ProcessingTask, int64, error) {
	return s.Query.ListTasks(ctx, workflowID, status, page, pageSize)
}

func (s *DataProcessingService) ListWorkflows(ctx context.Context, page, pageSize int) ([]*domain.ProcessingWorkflow, int64, error) {
	return s.Query.ListWorkflows(ctx, page, pageSize)
}
