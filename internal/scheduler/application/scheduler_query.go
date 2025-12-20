package application

import (
	"context"

	"github.com/wyfcoding/ecommerce/internal/scheduler/domain"
)

// SchedulerQuery 处理调度任务和日志的读操作。
type SchedulerQuery struct {
	repo domain.SchedulerRepository
}

// NewSchedulerQuery creates a new SchedulerQuery instance.
func NewSchedulerQuery(repo domain.SchedulerRepository) *SchedulerQuery {
	return &SchedulerQuery{
		repo: repo,
	}
}

// ListJobs 获取定时任务列表。
func (q *SchedulerQuery) ListJobs(ctx context.Context, status *int, page, pageSize int) ([]*domain.Job, int64, error) {
	offset := (page - 1) * pageSize
	var st *domain.JobStatus
	if status != nil {
		s := domain.JobStatus(*status)
		st = &s
	}
	return q.repo.ListJobs(ctx, st, offset, pageSize)
}

// ListJobLogs 获取任务日志列表。
func (q *SchedulerQuery) ListJobLogs(ctx context.Context, jobID uint64, page, pageSize int) ([]*domain.JobLog, int64, error) {
	offset := (page - 1) * pageSize
	return q.repo.ListJobLogs(ctx, jobID, offset, pageSize)
}

// GetJob 获取单个任务详情
func (q *SchedulerQuery) GetJob(ctx context.Context, id uint64) (*domain.Job, error) {
	return q.repo.GetJob(ctx, id)
}
