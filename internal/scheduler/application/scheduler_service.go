package application

import (
	"context"

	"github.com/wyfcoding/ecommerce/internal/scheduler/domain"
)

// SchedulerService 作为调度操作的门面。
type SchedulerService struct {
	manager *SchedulerManager
	query   *SchedulerQuery
}

// NewSchedulerService creates a new SchedulerService facade.
func NewSchedulerService(manager *SchedulerManager, query *SchedulerQuery) *SchedulerService {
	return &SchedulerService{
		manager: manager,
		query:   query,
	}
}

// --- 写操作（委托给 Manager）---

func (s *SchedulerService) CreateJob(ctx context.Context, name, desc, cron, handler, params string) (*domain.Job, error) {
	return s.manager.CreateJob(ctx, name, desc, cron, handler, params)
}

func (s *SchedulerService) UpdateJob(ctx context.Context, id uint64, cron, params string) error {
	return s.manager.UpdateJob(ctx, id, cron, params)
}

func (s *SchedulerService) ToggleJobStatus(ctx context.Context, id uint64, enable bool) error {
	return s.manager.ToggleJobStatus(ctx, id, enable)
}

func (s *SchedulerService) RunJob(ctx context.Context, id uint64) error {
	return s.manager.RunJob(ctx, id)
}

// --- 读操作（委托给 Query）---

func (s *SchedulerService) ListJobs(ctx context.Context, status *int, page, pageSize int) ([]*domain.Job, int64, error) {
	return s.query.ListJobs(ctx, status, page, pageSize)
}

func (s *SchedulerService) ListJobLogs(ctx context.Context, jobID uint64, page, pageSize int) ([]*domain.JobLog, int64, error) {
	return s.query.ListJobLogs(ctx, jobID, page, pageSize)
}

func (s *SchedulerService) GetJob(ctx context.Context, id uint64) (*domain.Job, error) {
	return s.query.GetJob(ctx, id)
}
