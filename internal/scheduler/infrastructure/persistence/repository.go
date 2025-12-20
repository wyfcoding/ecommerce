package persistence

import (
	"context"
	"errors"

	"github.com/wyfcoding/ecommerce/internal/scheduler/domain"
	"gorm.io/gorm"
)

type schedulerRepository struct {
	db *gorm.DB
}

// NewSchedulerRepository 创建并返回一个新的 schedulerRepository 实例。
func NewSchedulerRepository(db *gorm.DB) domain.SchedulerRepository {
	return &schedulerRepository{db: db}
}

// --- 任务管理 (Job methods) ---

func (r *schedulerRepository) SaveJob(ctx context.Context, job *domain.Job) error {
	return r.db.WithContext(ctx).Save(job).Error
}

func (r *schedulerRepository) GetJob(ctx context.Context, id uint64) (*domain.Job, error) {
	var job domain.Job
	if err := r.db.WithContext(ctx).First(&job, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &job, nil
}

func (r *schedulerRepository) GetJobByName(ctx context.Context, name string) (*domain.Job, error) {
	var job domain.Job
	if err := r.db.WithContext(ctx).Where("name = ?", name).First(&job).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &job, nil
}

func (r *schedulerRepository) ListJobs(ctx context.Context, status *domain.JobStatus, offset, limit int) ([]*domain.Job, int64, error) {
	var list []*domain.Job
	var total int64

	db := r.db.WithContext(ctx).Model(&domain.Job{})
	if status != nil {
		db = db.Where("status = ?", *status)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := db.Offset(offset).Limit(limit).Order("id desc").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

func (r *schedulerRepository) DeleteJob(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&domain.Job{}, id).Error
}

// --- 日志管理 (JobLog methods) ---

func (r *schedulerRepository) SaveJobLog(ctx context.Context, log *domain.JobLog) error {
	return r.db.WithContext(ctx).Save(log).Error
}

func (r *schedulerRepository) GetJobLog(ctx context.Context, id uint64) (*domain.JobLog, error) {
	var log domain.JobLog
	if err := r.db.WithContext(ctx).First(&log, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &log, nil
}

func (r *schedulerRepository) ListJobLogs(ctx context.Context, jobID uint64, offset, limit int) ([]*domain.JobLog, int64, error) {
	var list []*domain.JobLog
	var total int64

	db := r.db.WithContext(ctx).Model(&domain.JobLog{})
	if jobID > 0 {
		db = db.Where("job_id = ?", jobID)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := db.Offset(offset).Limit(limit).Order("start_time desc").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}
