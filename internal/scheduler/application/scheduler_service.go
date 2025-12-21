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

// NewSchedulerService 创建调度服务门面实例。
func NewSchedulerService(manager *SchedulerManager, query *SchedulerQuery) *SchedulerService {
	return &SchedulerService{
		manager: manager,
		query:   query,
	}
}

// --- 写操作（委托给 Manager）---

// CreateJob 创建一个新的定时任务。
func (s *SchedulerService) CreateJob(ctx context.Context, name, desc, cron, handler, params string) (*domain.Job, error) {
	return s.manager.CreateJob(ctx, name, desc, cron, handler, params)
}

// UpdateJob 更新现有定时任务的调度周期或参数。
func (s *SchedulerService) UpdateJob(ctx context.Context, id uint64, cron, params string) error {
	return s.manager.UpdateJob(ctx, id, cron, params)
}

// ToggleJobStatus 启用或停用指定的定时任务。
func (s *SchedulerService) ToggleJobStatus(ctx context.Context, id uint64, enable bool) error {
	return s.manager.ToggleJobStatus(ctx, id, enable)
}

// RunJob 立即手动执行一次指定的定时任务。
func (s *SchedulerService) RunJob(ctx context.Context, id uint64) error {
	return s.manager.RunJob(ctx, id)
}

// --- 读操作（委托给 Query）---

// ListJobs 分页获取定时任务列表。
func (s *SchedulerService) ListJobs(ctx context.Context, status *int, page, pageSize int) ([]*domain.Job, int64, error) {
	return s.query.ListJobs(ctx, status, page, pageSize)
}

// ListJobLogs 分页获取指定任务的执行历史日志。
func (s *SchedulerService) ListJobLogs(ctx context.Context, jobID uint64, page, pageSize int) ([]*domain.JobLog, int64, error) {
	return s.query.ListJobLogs(ctx, jobID, page, pageSize)
}

// GetJob 获取指定ID的定时任务配置详情。
func (s *SchedulerService) GetJob(ctx context.Context, id uint64) (*domain.Job, error) {
	return s.query.GetJob(ctx, id)
}
