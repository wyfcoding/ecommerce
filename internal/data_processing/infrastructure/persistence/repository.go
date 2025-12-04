package persistence

import (
	"context"
	"github.com/wyfcoding/ecommerce/internal/data_processing/domain/entity"     // 导入数据处理模块的领域实体定义。
	"github.com/wyfcoding/ecommerce/internal/data_processing/domain/repository" // 导入数据处理模块的领域仓储接口。

	"gorm.io/gorm" // 导入GORM ORM框架。
)

// dataProcessingRepository 是 DataProcessingRepository 接口的GORM实现。
// 它负责将数据处理模块的领域实体映射到数据库，并执行持久化操作。
type dataProcessingRepository struct {
	db *gorm.DB // GORM数据库连接实例。
}

// NewDataProcessingRepository 创建并返回一个新的 dataProcessingRepository 实例。
// db: GORM数据库连接实例。
func NewDataProcessingRepository(db *gorm.DB) repository.DataProcessingRepository {
	return &dataProcessingRepository{db: db}
}

// --- Task methods ---

// SaveTask 将处理任务实体保存到数据库。
// 如果任务已存在（通过ID），则更新其信息；如果不存在，则创建。
func (r *dataProcessingRepository) SaveTask(ctx context.Context, task *entity.ProcessingTask) error {
	return r.db.WithContext(ctx).Save(task).Error
}

// GetTask 根据ID从数据库获取处理任务记录。
func (r *dataProcessingRepository) GetTask(ctx context.Context, id uint64) (*entity.ProcessingTask, error) {
	var task entity.ProcessingTask
	if err := r.db.WithContext(ctx).First(&task, id).Error; err != nil {
		return nil, err
	}
	return &task, nil
}

// ListTasks 从数据库列出所有处理任务记录，支持通过工作流ID和状态过滤，并支持分页。
func (r *dataProcessingRepository) ListTasks(ctx context.Context, workflowID uint64, status entity.TaskStatus, offset, limit int) ([]*entity.ProcessingTask, int64, error) {
	var list []*entity.ProcessingTask
	var total int64

	db := r.db.WithContext(ctx).Model(&entity.ProcessingTask{})
	if workflowID != 0 { // 如果提供了工作流ID，则按工作流ID过滤。
		db = db.Where("workflow_id = ?", workflowID)
	}
	if status != 0 { // 如果提供了任务状态，则按状态过滤。
		db = db.Where("status = ?", status)
	}

	// 统计总记录数。
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 应用分页和排序。
	if err := db.Offset(offset).Limit(limit).Order("created_at desc").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// UpdateTask 更新数据库中的处理任务记录。
func (r *dataProcessingRepository) UpdateTask(ctx context.Context, task *entity.ProcessingTask) error {
	return r.db.WithContext(ctx).Save(task).Error
}

// --- Workflow methods ---

// SaveWorkflow 将工作流实体保存到数据库。
// 如果工作流已存在（通过ID），则更新其信息；如果不存在，则创建。
func (r *dataProcessingRepository) SaveWorkflow(ctx context.Context, workflow *entity.ProcessingWorkflow) error {
	return r.db.WithContext(ctx).Save(workflow).Error
}

// GetWorkflow 根据ID从数据库获取工作流记录。
func (r *dataProcessingRepository) GetWorkflow(ctx context.Context, id uint64) (*entity.ProcessingWorkflow, error) {
	var workflow entity.ProcessingWorkflow
	if err := r.db.WithContext(ctx).First(&workflow, id).Error; err != nil {
		return nil, err
	}
	return &workflow, nil
}

// ListWorkflows 从数据库列出所有工作流记录，支持分页。
func (r *dataProcessingRepository) ListWorkflows(ctx context.Context, offset, limit int) ([]*entity.ProcessingWorkflow, int64, error) {
	var list []*entity.ProcessingWorkflow
	var total int64

	db := r.db.WithContext(ctx).Model(&entity.ProcessingWorkflow{})

	// 统计总记录数。
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 应用分页和排序。
	if err := db.Offset(offset).Limit(limit).Order("created_at desc").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// UpdateWorkflow 更新数据库中的工作流记录。
func (r *dataProcessingRepository) UpdateWorkflow(ctx context.Context, workflow *entity.ProcessingWorkflow) error {
	return r.db.WithContext(ctx).Save(workflow).Error
}
