package repository

import (
	"context"
	"github.com/wyfcoding/ecommerce/internal/data_processing/domain/entity" // 导入数据处理领域的实体定义。
)

// DataProcessingRepository 是数据处理模块的仓储接口。
// 它定义了对处理任务和工作流实体进行数据持久化操作的契约。
// 仓储接口属于领域层，旨在将领域逻辑与数据存储的实现细节解耦。
type DataProcessingRepository interface {
	// --- Task methods ---

	// SaveTask 将处理任务实体保存到数据存储中。
	// 如果任务已存在，则更新；如果不存在，则创建。
	// ctx: 上下文。
	// task: 待保存的任务实体。
	SaveTask(ctx context.Context, task *entity.ProcessingTask) error
	// GetTask 根据ID获取处理任务实体。
	GetTask(ctx context.Context, id uint64) (*entity.ProcessingTask, error)
	// ListTasks 列出指定工作流ID和状态的所有处理任务实体，支持分页。
	ListTasks(ctx context.Context, workflowID uint64, status entity.TaskStatus, offset, limit int) ([]*entity.ProcessingTask, int64, error)
	// UpdateTask 更新处理任务实体的信息。
	UpdateTask(ctx context.Context, task *entity.ProcessingTask) error

	// --- Workflow methods ---

	// SaveWorkflow 将工作流实体保存到数据存储中。
	SaveWorkflow(ctx context.Context, workflow *entity.ProcessingWorkflow) error
	// GetWorkflow 根据ID获取工作流实体。
	GetWorkflow(ctx context.Context, id uint64) (*entity.ProcessingWorkflow, error)
	// ListWorkflows 列出所有工作流实体，支持分页。
	ListWorkflows(ctx context.Context, offset, limit int) ([]*entity.ProcessingWorkflow, int64, error)
	// UpdateWorkflow 更新工作流实体的信息。
	UpdateWorkflow(ctx context.Context, workflow *entity.ProcessingWorkflow) error
}
