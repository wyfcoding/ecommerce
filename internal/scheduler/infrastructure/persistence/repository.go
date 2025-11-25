package persistence

import (
	"context"
	"github.com/wyfcoding/ecommerce/internal/scheduler/domain/entity"
	"github.com/wyfcoding/ecommerce/internal/scheduler/domain/repository"
	"errors"

	"gorm.io/gorm"
)

type schedulerRepository struct {
	db *gorm.DB
}

func NewSchedulerRepository(db *gorm.DB) repository.SchedulerRepository {
	return &schedulerRepository{db: db}
}

// 任务管理
func (r *schedulerRepository) SaveJob(ctx context.Context, job *entity.Job) error {
	return r.db.WithContext(ctx).Save(job).Error
}

func (r *schedulerRepository) GetJob(ctx context.Context, id uint64) (*entity.Job, error) {
	var job entity.Job
	if err := r.db.WithContext(ctx).First(&job, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &job, nil
}

func (r *schedulerRepository) GetJobByName(ctx context.Context, name string) (*entity.Job, error) {
	var job entity.Job
	if err := r.db.WithContext(ctx).Where("name = ?", name).First(&job).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &job, nil
}

func (r *schedulerRepository) ListJobs(ctx context.Context, status *entity.JobStatus, offset, limit int) ([]*entity.Job, int64, error) {
	var list []*entity.Job
	var total int64

	db := r.db.WithContext(ctx).Model(&entity.Job{})
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
	return r.db.WithContext(ctx).Delete(&entity.Job{}, id).Error
}

// 日志管理
func (r *schedulerRepository) SaveJobLog(ctx context.Context, log *entity.JobLog) error {
	return r.db.WithContext(ctx).Save(log).Error
}

func (r *schedulerRepository) GetJobLog(ctx context.Context, id uint64) (*entity.JobLog, error) {
	var log entity.JobLog
	if err := r.db.WithContext(ctx).First(&log, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &log, nil
}

func (r *schedulerRepository) ListJobLogs(ctx context.Context, jobID uint64, offset, limit int) ([]*entity.JobLog, int64, error) {
	var list []*entity.JobLog
	var total int64

	db := r.db.WithContext(ctx).Model(&entity.JobLog{})
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
