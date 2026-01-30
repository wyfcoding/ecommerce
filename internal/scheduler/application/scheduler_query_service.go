package application

import (
	"context"

	"github.com/wyfcoding/ecommerce/internal/scheduler/domain"
)

// SchedulerQueryService 处理调度任务和日志的读操作。
type SchedulerQueryService struct {
	repo domain.SchedulerRepository
}

// NewSchedulerQueryService creates a new SchedulerQueryService instance.
func NewSchedulerQueryService(repo domain.SchedulerRepository) *SchedulerQueryService {
	return &SchedulerQueryService{
		repo: repo,
	}
}

// ListJobs 获取定时任务列表。
func (q *SchedulerQueryService) ListJobs(ctx context.Context, status *int, page, pageSize int) ([]*domain.Job, int64, error) {
	offset := (page - 1) * pageSize
	var st *domain.JobStatus
	if status != nil {
		s := domain.JobStatus(*status)
		st = &s
	}
	return q.repo.ListJobs(ctx, st, offset, pageSize)
}

// ListJobLogs 获取任务日志列表。
func (q *SchedulerQueryService) ListJobLogs(ctx context.Context, jobID uint64, page, pageSize int) ([]*domain.JobLog, int64, error) {
	offset := (page - 1) * pageSize
	return q.repo.ListJobLogs(ctx, jobID, offset, pageSize)
}

// GetJob 获取单个任务详情
func (q *SchedulerQueryService) GetJob(ctx context.Context, id uint64) (*domain.Job, error) {
	return q.repo.GetJob(ctx, id)
}
