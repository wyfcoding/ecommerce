package persistence

import (
	"context"
	"ecommerce/internal/data_processing/domain/entity"
	"ecommerce/internal/data_processing/domain/repository"

	"gorm.io/gorm"
)

type dataProcessingRepository struct {
	db *gorm.DB
}

func NewDataProcessingRepository(db *gorm.DB) repository.DataProcessingRepository {
	return &dataProcessingRepository{db: db}
}

// Task methods
func (r *dataProcessingRepository) SaveTask(ctx context.Context, task *entity.ProcessingTask) error {
	return r.db.WithContext(ctx).Save(task).Error
}

func (r *dataProcessingRepository) GetTask(ctx context.Context, id uint64) (*entity.ProcessingTask, error) {
	var task entity.ProcessingTask
	if err := r.db.WithContext(ctx).First(&task, id).Error; err != nil {
		return nil, err
	}
	return &task, nil
}

func (r *dataProcessingRepository) ListTasks(ctx context.Context, workflowID uint64, status entity.TaskStatus, offset, limit int) ([]*entity.ProcessingTask, int64, error) {
	var list []*entity.ProcessingTask
	var total int64

	db := r.db.WithContext(ctx).Model(&entity.ProcessingTask{})
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

func (r *dataProcessingRepository) UpdateTask(ctx context.Context, task *entity.ProcessingTask) error {
	return r.db.WithContext(ctx).Save(task).Error
}

// Workflow methods
func (r *dataProcessingRepository) SaveWorkflow(ctx context.Context, workflow *entity.ProcessingWorkflow) error {
	return r.db.WithContext(ctx).Save(workflow).Error
}

func (r *dataProcessingRepository) GetWorkflow(ctx context.Context, id uint64) (*entity.ProcessingWorkflow, error) {
	var workflow entity.ProcessingWorkflow
	if err := r.db.WithContext(ctx).First(&workflow, id).Error; err != nil {
		return nil, err
	}
	return &workflow, nil
}

func (r *dataProcessingRepository) ListWorkflows(ctx context.Context, offset, limit int) ([]*entity.ProcessingWorkflow, int64, error) {
	var list []*entity.ProcessingWorkflow
	var total int64

	db := r.db.WithContext(ctx).Model(&entity.ProcessingWorkflow{})

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := db.Offset(offset).Limit(limit).Order("created_at desc").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

func (r *dataProcessingRepository) UpdateWorkflow(ctx context.Context, workflow *entity.ProcessingWorkflow) error {
	return r.db.WithContext(ctx).Save(workflow).Error
}
