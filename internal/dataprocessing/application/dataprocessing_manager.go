package application

import (
	"context"
	"log/slog"
	"time"

	"github.com/wyfcoding/ecommerce/internal/dataprocessing/domain"
)

// DataProcessingManager 处理所有数据处理相关的写入操作（Commands）。
type DataProcessingManager struct {
	repo   domain.DataProcessingRepository
	logger *slog.Logger
}

// NewDataProcessingManager 构造函数。
func NewDataProcessingManager(repo domain.DataProcessingRepository, logger *slog.Logger) *DataProcessingManager {
	return &DataProcessingManager{
		repo:   repo,
		logger: logger,
	}
}

// SubmitTask 提交一个数据处理任务。
func (m *DataProcessingManager) SubmitTask(ctx context.Context, name, taskType, config string, workflowID uint64) (*domain.ProcessingTask, error) {
	task := domain.NewProcessingTask(name, taskType, config, workflowID)
	if err := m.repo.SaveTask(ctx, task); err != nil {
		m.logger.ErrorContext(ctx, "failed to save task", "name", name, "error", err)
		return nil, err
	}
	m.logger.InfoContext(ctx, "task submitted successfully", "task_id", task.ID, "name", name)

	// 异步处理任务。
	go m.processTask(task)

	return task, nil
}

// processTask 异步处理数据处理任务的后台逻辑。
func (m *DataProcessingManager) processTask(task *domain.ProcessingTask) {
	ctx := context.Background()
	task.Start()
	if err := m.repo.UpdateTask(ctx, task); err != nil {
		m.logger.ErrorContext(ctx, "failed to update task status to running", "task_id", task.ID, "error", err)
		return
	}

	// 模拟: 模拟数据处理过程。
	time.Sleep(1 * time.Second)

	// Success simulation
	task.Complete(`{"status": "success", "data": "processed"}`)
	if err := m.repo.UpdateTask(ctx, task); err != nil {
		m.logger.ErrorContext(ctx, "failed to update task status to completed", "task_id", task.ID, "error", err)
	}
}

// CreateWorkflow 创建一个新的数据处理工作流。
func (m *DataProcessingManager) CreateWorkflow(ctx context.Context, name, description, steps string) (*domain.ProcessingWorkflow, error) {
	workflow := domain.NewProcessingWorkflow(name, description, steps)
	if err := m.repo.SaveWorkflow(ctx, workflow); err != nil {
		m.logger.ErrorContext(ctx, "failed to save workflow", "name", name, "error", err)
		return nil, err
	}
	m.logger.InfoContext(ctx, "workflow created successfully", "workflow_id", workflow.ID, "name", name)
	return workflow, nil
}
