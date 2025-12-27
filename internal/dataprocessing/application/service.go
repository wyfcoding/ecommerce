package application

import (
	"context"
	"log/slog"
	"time"

	"github.com/wyfcoding/ecommerce/internal/dataprocessing/domain"
)

// DataProcessingService 结构体定义了数据处理相关的应用服务。
type DataProcessingService struct {
	repo   domain.DataProcessingRepository
	logger *slog.Logger
}

// NewDataProcessingService 创建并返回一个新的 DataProcessingService 实例。
func NewDataProcessingService(repo domain.DataProcessingRepository, logger *slog.Logger) *DataProcessingService {
	return &DataProcessingService{
		repo:   repo,
		logger: logger,
	}
}

// SubmitTask 提交一个数据处理任务。
func (s *DataProcessingService) SubmitTask(ctx context.Context, name, taskType, config string, workflowID uint64) (*domain.ProcessingTask, error) {
	task := domain.NewProcessingTask(name, taskType, config, workflowID)
	if err := s.repo.SaveTask(ctx, task); err != nil {
		s.logger.ErrorContext(ctx, "failed to save task", "name", name, "error", err)
		return nil, err
	}
	s.logger.InfoContext(ctx, "task submitted successfully", "task_id", task.ID, "name", name)

	// 异步处理任务。
	go s.processTask(task)

	return task, nil
}

// processTask 异步处理数据处理任务的后台逻辑。
func (s *DataProcessingService) processTask(task *domain.ProcessingTask) {
	ctx := context.Background()
	task.Start()
	if err := s.repo.UpdateTask(ctx, task); err != nil {
		s.logger.ErrorContext(ctx, "failed to update task status to running", "task_id", task.ID, "error", err)
		return
	}

	// 模拟: 模拟数据处理过程。
	time.Sleep(1 * time.Second)

	// Success simulation
	task.Complete(`{"status": "success", "data": "processed"}`)
	if err := s.repo.UpdateTask(ctx, task); err != nil {
		s.logger.ErrorContext(ctx, "failed to update task status to completed", "task_id", task.ID, "error", err)
	}
}

// ListTasks 获取数据处理任务列表。
func (s *DataProcessingService) ListTasks(ctx context.Context, workflowID uint64, status domain.TaskStatus, page, pageSize int) ([]*domain.ProcessingTask, int64, error) {
	offset := (page - 1) * pageSize
	return s.repo.ListTasks(ctx, workflowID, status, offset, pageSize)
}

// CreateWorkflow 创建一个新的数据处理工作流。
func (s *DataProcessingService) CreateWorkflow(ctx context.Context, name, description, steps string) (*domain.ProcessingWorkflow, error) {
	workflow := domain.NewProcessingWorkflow(name, description, steps)
	if err := s.repo.SaveWorkflow(ctx, workflow); err != nil {
		s.logger.ErrorContext(ctx, "failed to save workflow", "name", name, "error", err)
		return nil, err
	}
	s.logger.InfoContext(ctx, "workflow created successfully", "workflow_id", workflow.ID, "name", name)
	return workflow, nil
}

// ListWorkflows 获取数据处理工作流列表。
func (s *DataProcessingService) ListWorkflows(ctx context.Context, page, pageSize int) ([]*domain.ProcessingWorkflow, int64, error) {
	offset := (page - 1) * pageSize
	return s.repo.ListWorkflows(ctx, offset, pageSize)
}
