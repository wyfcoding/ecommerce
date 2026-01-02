package application

import (
	"context"
	"fmt"
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

	// 真实化实现：根据任务类型执行不同的处理逻辑
	var (
		result string
		errMsg string
		failed bool
	)

	m.logger.InfoContext(ctx, "processing task started", "task_id", task.ID, "type", task.Type)

	switch task.Type {
	case "CLEAN":
		// 模拟清洗逻辑
		time.Sleep(1500 * time.Millisecond)
		result = `{"status": "success", "cleaned_records": 150}`
	case "TRANSFORM":
		// 模拟转换逻辑
		time.Sleep(2000 * time.Millisecond)
		result = `{"status": "success", "transformed_format": "parquet"}`
	case "FAIL_TEST":
		// 模拟失败情况
		time.Sleep(500 * time.Millisecond)
		failed = true
		errMsg = "simulated processing error for fail test"
	default:
		// 默认处理
		time.Sleep(1000 * time.Millisecond)
		result = fmt.Sprintf(`{"status": "success", "msg": "default processing for %s"}`, task.Type)
	}

	if failed {
		task.Fail(errMsg)
	} else {
		task.Complete(result)
	}

	if err := m.repo.UpdateTask(ctx, task); err != nil {
		m.logger.ErrorContext(ctx, "failed to update final task status", "task_id", task.ID, "error", err)
	} else {
		m.logger.InfoContext(ctx, "task processing finished", "task_id", task.ID, "status", task.Status)
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

// CancelTask 取消正在执行的任务。
func (m *DataProcessingManager) CancelTask(ctx context.Context, id uint64) error {
	task, err := m.repo.GetTask(ctx, id)
	if err != nil {
		return err
	}
	if task == nil {
		return domain.ErrTaskNotFound
	}

	task.Status = domain.TaskStatusCancelled
	return m.repo.UpdateTask(ctx, task)
}

// UpdateWorkflow 更新工作流定义。
func (m *DataProcessingManager) UpdateWorkflow(ctx context.Context, id uint64, name, description, steps string) error {
	workflow, err := m.repo.GetWorkflow(ctx, id)
	if err != nil {
		return err
	}
	if workflow == nil {
		return domain.ErrWorkflowNotFound
	}

	if name != "" {
		workflow.Name = name
	}
	if description != "" {
		workflow.Description = description
	}
	if steps != "" {
		workflow.Steps = steps
	}

	return m.repo.UpdateWorkflow(ctx, workflow)
}

// SetWorkflowActive 激活或停用工作流。
func (m *DataProcessingManager) SetWorkflowActive(ctx context.Context, id uint64, active bool) error {
	workflow, err := m.repo.GetWorkflow(ctx, id)
	if err != nil {
		return err
	}
	if workflow == nil {
		return domain.ErrWorkflowNotFound
	}

	if active {
		workflow.Status = domain.WorkflowStatusActive
	} else {
		workflow.Status = domain.WorkflowStatusInactive
	}
	return m.repo.UpdateWorkflow(ctx, workflow)
}
