package application

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/wyfcoding/ecommerce/internal/scheduler/domain"
)

// SchedulerManager 处理调度任务和日志的写操作。
type SchedulerManager struct {
	repo   domain.SchedulerRepository
	logger *slog.Logger
}

// NewSchedulerManager creates a new SchedulerManager instance.
func NewSchedulerManager(repo domain.SchedulerRepository, logger *slog.Logger) *SchedulerManager {
	return &SchedulerManager{
		repo:   repo,
		logger: logger,
	}
}

// CreateJob 创建一个新的定时任务。
func (m *SchedulerManager) CreateJob(ctx context.Context, name, desc, cron, handler, params string) (*domain.Job, error) {
	existing, err := m.repo.GetJobByName(ctx, name)
	if err != nil {
		m.logger.ErrorContext(ctx, "failed to check existing job name", "job_name", name, "error", err)
		return nil, err
	}
	if existing != nil {
		return nil, errors.New("job name already exists")
	}

	job := &domain.Job{
		Name:        name,
		Description: desc,
		CronExpr:    cron,
		Handler:     handler,
		Params:      params,
		Status:      domain.JobStatusEnabled,
	}

	if err := m.repo.SaveJob(ctx, job); err != nil {
		m.logger.ErrorContext(ctx, "failed to save job", "job_name", name, "error", err)
		return nil, err
	}
	m.logger.InfoContext(ctx, "job created successfully", "job_id", job.ID, "job_name", name)
	return job, nil
}

// UpdateJob 更新指定ID的定时任务信息。
func (m *SchedulerManager) UpdateJob(ctx context.Context, id uint64, cron, params string) error {
	job, err := m.repo.GetJob(ctx, id)
	if err != nil {
		return err
	}
	if job == nil {
		return errors.New("job not found")
	}

	job.CronExpr = cron
	job.Params = params

	if err := m.repo.SaveJob(ctx, job); err != nil {
		m.logger.ErrorContext(ctx, "failed to update job", "job_id", id, "error", err)
		return err
	}
	m.logger.InfoContext(ctx, "job updated successfully", "job_id", id)
	return nil
}

// ToggleJobStatus 切换定时任务的启用/禁用状态。
func (m *SchedulerManager) ToggleJobStatus(ctx context.Context, id uint64, enable bool) error {
	job, err := m.repo.GetJob(ctx, id)
	if err != nil {
		return err
	}
	if job == nil {
		return errors.New("job not found")
	}

	if enable {
		job.Status = domain.JobStatusEnabled
	} else {
		job.Status = domain.JobStatusDisabled
	}

	if err := m.repo.SaveJob(ctx, job); err != nil {
		m.logger.ErrorContext(ctx, "failed to toggle job status", "job_id", id, "enable", enable, "error", err)
		return err
	}
	m.logger.InfoContext(ctx, "job status toggled successfully", "job_id", id, "enable", enable)
	return nil
}

// RunJob 立即运行指定ID的定时任务。
func (m *SchedulerManager) RunJob(ctx context.Context, id uint64) error {
	job, err := m.repo.GetJob(ctx, id)
	if err != nil {
		return err
	}
	if job == nil {
		return errors.New("job not found")
	}

	if job.Status == domain.JobStatusRunning {
		return errors.New("job is already running")
	}

	log := &domain.JobLog{
		JobID:     uint64(job.ID),
		JobName:   job.Name,
		Handler:   job.Handler,
		Params:    job.Params,
		Status:    "RUNNING",
		StartTime: time.Now(),
	}
	if err := m.repo.SaveJobLog(ctx, log); err != nil {
		m.logger.ErrorContext(ctx, "failed to save job log", "job_id", id, "error", err)
		return err
	}

	job.Status = domain.JobStatusRunning
	now := time.Now()
	job.LastRunTime = &now
	job.RunCount++
	if err := m.repo.SaveJob(ctx, job); err != nil {
		m.logger.ErrorContext(ctx, "failed to update job status to running", "job_id", id, "error", err)
		return err
	}
	m.logger.InfoContext(ctx, "job started execution (simulated)", "job_id", id)

	go func() {
		// Simulate job execution
		time.Sleep(1 * time.Second)

		endTime := time.Now()
		log.EndTime = &endTime
		log.Duration = endTime.Sub(log.StartTime).Milliseconds()
		log.Status = "SUCCESS"
		log.Result = "Executed successfully"

		_ = m.repo.SaveJobLog(context.Background(), log)
		m.logger.Info("job execution completed (simulated)", "job_id", id, "status", log.Status)

		job.Status = domain.JobStatusEnabled
		_ = m.repo.SaveJob(context.Background(), job)
		m.logger.Info("job status reset to enabled (simulated)", "job_id", id)
	}()

	return nil
}
