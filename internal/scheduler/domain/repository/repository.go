package repository

import (
	"context"
	"ecommerce/internal/scheduler/domain/entity"
)

// SchedulerRepository 调度器仓储接口
type SchedulerRepository interface {
	// 任务管理
	SaveJob(ctx context.Context, job *entity.Job) error
	GetJob(ctx context.Context, id uint64) (*entity.Job, error)
	GetJobByName(ctx context.Context, name string) (*entity.Job, error)
	ListJobs(ctx context.Context, status *entity.JobStatus, offset, limit int) ([]*entity.Job, int64, error)
	DeleteJob(ctx context.Context, id uint64) error

	// 日志管理
	SaveJobLog(ctx context.Context, log *entity.JobLog) error
	GetJobLog(ctx context.Context, id uint64) (*entity.JobLog, error)
	ListJobLogs(ctx context.Context, jobID uint64, offset, limit int) ([]*entity.JobLog, int64, error)
}
