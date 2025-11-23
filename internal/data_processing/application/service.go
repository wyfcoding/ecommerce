package application

import (
	"context"
	"ecommerce/internal/data_processing/domain/entity"
	"ecommerce/internal/data_processing/domain/repository"
	"time"

	"log/slog"
)

type DataProcessingService struct {
	repo   repository.DataProcessingRepository
	logger *slog.Logger
}

func NewDataProcessingService(repo repository.DataProcessingRepository, logger *slog.Logger) *DataProcessingService {
	return &DataProcessingService{
		repo:   repo,
		logger: logger,
	}
}

// SubmitTask 提交任务
func (s *DataProcessingService) SubmitTask(ctx context.Context, name, taskType, config string, workflowID uint64) (*entity.ProcessingTask, error) {
	task := entity.NewProcessingTask(name, taskType, config, workflowID)
	if err := s.repo.SaveTask(ctx, task); err != nil {
		s.logger.Error("failed to save task", "error", err)
		return nil, err
	}

	// Async processing simulation
	go s.processTask(task)

	return task, nil
}

func (s *DataProcessingService) processTask(task *entity.ProcessingTask) {
	ctx := context.Background()
	task.Start()
	s.repo.UpdateTask(ctx, task)

	// Simulate processing
	time.Sleep(1 * time.Second)

	// Success simulation
	task.Complete(`{"status": "success", "data": "processed"}`)
	s.repo.UpdateTask(ctx, task)
}

// ListTasks 获取任务列表
func (s *DataProcessingService) ListTasks(ctx context.Context, workflowID uint64, status entity.TaskStatus, page, pageSize int) ([]*entity.ProcessingTask, int64, error) {
	offset := (page - 1) * pageSize
	return s.repo.ListTasks(ctx, workflowID, status, offset, pageSize)
}

// CreateWorkflow 创建工作流
func (s *DataProcessingService) CreateWorkflow(ctx context.Context, name, description, steps string) (*entity.ProcessingWorkflow, error) {
	workflow := entity.NewProcessingWorkflow(name, description, steps)
	if err := s.repo.SaveWorkflow(ctx, workflow); err != nil {
		s.logger.Error("failed to save workflow", "error", err)
		return nil, err
	}
	return workflow, nil
}

// ListWorkflows 获取工作流列表
func (s *DataProcessingService) ListWorkflows(ctx context.Context, page, pageSize int) ([]*entity.ProcessingWorkflow, int64, error) {
	offset := (page - 1) * pageSize
	return s.repo.ListWorkflows(ctx, offset, pageSize)
}
