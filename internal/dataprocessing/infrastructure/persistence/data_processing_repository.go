package persistence

import (
	"context"

	"github.com/wyfcoding/ecommerce/internal/dataprocessing/domain"
	"gorm.io/gorm"
)

type dataProcessingRepository struct {
	db *gorm.DB
}

// NewDataProcessingRepository 创建并返回一个新的 dataProcessingRepository 实例。
func NewDataProcessingRepository(db *gorm.DB) domain.DataProcessingRepository {
	return &dataProcessingRepository{db: db}
}

// --- Task methods ---

func (r *dataProcessingRepository) SaveTask(ctx context.Context, task *domain.ProcessingTask) error {
	return r.db.WithContext(ctx).Save(task).Error
}

func (r *dataProcessingRepository) GetTask(ctx context.Context, id uint64) (*domain.ProcessingTask, error) {
	var task domain.ProcessingTask
	if err := r.db.WithContext(ctx).First(&task, id).Error; err != nil {
		return nil, err
	}
	return &task, nil
}

func (r *dataProcessingRepository) ListTasks(ctx context.Context, workflowID uint64, status domain.TaskStatus, offset, limit int) ([]*domain.ProcessingTask, int64, error) {
	var list []*domain.ProcessingTask
	var total int64

	db := r.db.WithContext(ctx).Model(&domain.ProcessingTask{})
	if workflowID != 0 {
		db = db.Where("workflow_id = ?", workflowID)
	}
	if status != 0 {
		db = db.Where("status = ?", status)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := db.Offset(offset).Limit(limit).Order("created_at desc").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

func (r *dataProcessingRepository) UpdateTask(ctx context.Context, task *domain.ProcessingTask) error {
	return r.db.WithContext(ctx).Save(task).Error
}

// --- Workflow methods ---

func (r *dataProcessingRepository) SaveWorkflow(ctx context.Context, workflow *domain.ProcessingWorkflow) error {
	return r.db.WithContext(ctx).Save(workflow).Error
}

func (r *dataProcessingRepository) GetWorkflow(ctx context.Context, id uint64) (*domain.ProcessingWorkflow, error) {
	var workflow domain.ProcessingWorkflow
	if err := r.db.WithContext(ctx).First(&workflow, id).Error; err != nil {
		return nil, err
	}
	return &workflow, nil
}

func (r *dataProcessingRepository) ListWorkflows(ctx context.Context, offset, limit int) ([]*domain.ProcessingWorkflow, int64, error) {
	var list []*domain.ProcessingWorkflow
	var total int64

	db := r.db.WithContext(ctx).Model(&domain.ProcessingWorkflow{})

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := db.Offset(offset).Limit(limit).Order("created_at desc").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

func (r *dataProcessingRepository) UpdateWorkflow(ctx context.Context, workflow *domain.ProcessingWorkflow) error {
	return r.db.WithContext(ctx).Save(workflow).Error
}
