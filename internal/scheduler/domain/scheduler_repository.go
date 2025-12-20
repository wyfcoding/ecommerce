package domain

import (
	"context"
)

// SchedulerRepository 是调度模块的仓储接口。
type SchedulerRepository interface {
	// Job
	SaveJob(ctx context.Context, job *Job) error
	GetJob(ctx context.Context, id uint64) (*Job, error)
	GetJobByName(ctx context.Context, name string) (*Job, error)
	ListJobs(ctx context.Context, status *JobStatus, offset, limit int) ([]*Job, int64, error)
	DeleteJob(ctx context.Context, id uint64) error

	// JobLog
	SaveJobLog(ctx context.Context, log *JobLog) error
	GetJobLog(ctx context.Context, id uint64) (*JobLog, error)
	ListJobLogs(ctx context.Context, jobID uint64, offset, limit int) ([]*JobLog, int64, error)
}
