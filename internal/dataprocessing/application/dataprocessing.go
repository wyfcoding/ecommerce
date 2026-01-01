package application

import (
	"context"
	"log/slog"

	"github.com/wyfcoding/ecommerce/internal/dataprocessing/domain"
)

// DataProcessingService 数据处理门面服务，整合 Manager 和 Query。
type DataProcessingService struct {
	manager *DataProcessingManager
	query   *DataProcessingQuery
}

// NewDataProcessingService 构造函数。
func NewDataProcessingService(repo domain.DataProcessingRepository, logger *slog.Logger) *DataProcessingService {
	return &DataProcessingService{
		manager: NewDataProcessingManager(repo, logger),
		query:   NewDataProcessingQuery(repo),
	}
}

// --- Manager (Writes) ---

func (s *DataProcessingService) SubmitTask(ctx context.Context, name, taskType, config string, workflowID uint64) (*domain.ProcessingTask, error) {
	return s.manager.SubmitTask(ctx, name, taskType, config, workflowID)
}

func (s *DataProcessingService) CreateWorkflow(ctx context.Context, name, description, steps string) (*domain.ProcessingWorkflow, error) {
	return s.manager.CreateWorkflow(ctx, name, description, steps)
}

// --- Query (Reads) ---

func (s *DataProcessingService) ListTasks(ctx context.Context, workflowID uint64, status domain.TaskStatus, page, pageSize int) ([]*domain.ProcessingTask, int64, error) {
	return s.query.ListTasks(ctx, workflowID, status, page, pageSize)
}

func (s *DataProcessingService) ListWorkflows(ctx context.Context, page, pageSize int) ([]*domain.ProcessingWorkflow, int64, error) {
	return s.query.ListWorkflows(ctx, page, pageSize)
}
