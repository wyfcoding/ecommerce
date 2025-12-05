package application

import (
	"context"
	"time"

	"github.com/wyfcoding/ecommerce/internal/data_processing/domain/entity"     // 导入数据处理领域的实体定义。
	"github.com/wyfcoding/ecommerce/internal/data_processing/domain/repository" // 导入数据处理领域的仓储接口。

	"log/slog" // 导入结构化日志库。
)

// DataProcessingService 结构体定义了数据处理相关的应用服务。
// 它协调领域层和基础设施层，处理数据处理任务的提交、执行和工作流的管理等业务逻辑。
type DataProcessingService struct {
	repo   repository.DataProcessingRepository // 依赖DataProcessingRepository接口，用于数据持久化操作。
	logger *slog.Logger                        // 日志记录器，用于记录服务运行时的信息和错误。
}

// NewDataProcessingService 创建并返回一个新的 DataProcessingService 实例。
func NewDataProcessingService(repo repository.DataProcessingRepository, logger *slog.Logger) *DataProcessingService {
	return &DataProcessingService{
		repo:   repo,
		logger: logger,
	}
}

// SubmitTask 提交一个数据处理任务。
// ctx: 上下文。
// name: 任务名称。
// taskType: 任务类型。
// config: 任务配置（例如，JSON字符串）。
// workflowID: 关联的工作流ID。
// 返回created successfully的ProcessingTask实体和可能发生的错误。
func (s *DataProcessingService) SubmitTask(ctx context.Context, name, taskType, config string, workflowID uint64) (*entity.ProcessingTask, error) {
	task := entity.NewProcessingTask(name, taskType, config, workflowID) // 创建ProcessingTask实体。
	// 通过仓储接口保存任务。
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
// task: 待处理的任务实体。
func (s *DataProcessingService) processTask(task *entity.ProcessingTask) {
	ctx := context.Background()  // 使用一个新的背景上下文处理后台任务。
	task.Start()                 // 调用实体方法更新任务状态为运行中。
	s.repo.UpdateTask(ctx, task) // 更新数据库中的任务状态。

	// Simulate processing: 模拟数据处理过程。
	time.Sleep(1 * time.Second)

	// Success simulation: 模拟成功完成处理。
	task.Complete(`{"status": "success", "data": "processed"}`) // 调用实体方法更新任务状态为完成，并记录结果。
	s.repo.UpdateTask(ctx, task)                                // 更新数据库中的任务状态。
}

// ListTasks 获取数据处理任务列表。
// ctx: 上下文。
// workflowID: 筛选任务的工作流ID。
// status: 筛选任务的状态。
// page, pageSize: 分页参数。
// 返回任务列表、总数和可能发生的错误。
func (s *DataProcessingService) ListTasks(ctx context.Context, workflowID uint64, status entity.TaskStatus, page, pageSize int) ([]*entity.ProcessingTask, int64, error) {
	offset := (page - 1) * pageSize
	return s.repo.ListTasks(ctx, workflowID, status, offset, pageSize)
}

// CreateWorkflow 创建一个新的数据处理工作流。
// ctx: 上下文。
// name: 工作流名称。
// description: 工作流描述。
// steps: 工作流的步骤定义（例如，JSON或YAML字符串）。
// 返回created successfully的ProcessingWorkflow实体和可能发生的错误。
func (s *DataProcessingService) CreateWorkflow(ctx context.Context, name, description, steps string) (*entity.ProcessingWorkflow, error) {
	workflow := entity.NewProcessingWorkflow(name, description, steps) // 创建ProcessingWorkflow实体。
	// 通过仓储接口保存工作流。
	if err := s.repo.SaveWorkflow(ctx, workflow); err != nil {
		s.logger.ErrorContext(ctx, "failed to save workflow", "name", name, "error", err)
		return nil, err
	}
	s.logger.InfoContext(ctx, "workflow created successfully", "workflow_id", workflow.ID, "name", name)
	return workflow, nil
}

// ListWorkflows 获取数据处理工作流列表。
// ctx: 上下文。
// page, pageSize: 分页参数。
// 返回工作流列表、总数和可能发生的错误。
func (s *DataProcessingService) ListWorkflows(ctx context.Context, page, pageSize int) ([]*entity.ProcessingWorkflow, int64, error) {
	offset := (page - 1) * pageSize
	return s.repo.ListWorkflows(ctx, offset, pageSize)
}
