package application

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/wyfcoding/ecommerce/internal/scheduler/domain"
	"github.com/wyfcoding/pkg/algorithm"
)

// SchedulerManager 处理调度任务和日志的写操作。
// 引入时间轮（TimingWheel）用于管理海量的延迟任务（如订单超时取消）。
type SchedulerManager struct {
	repo       domain.SchedulerRepository
	logger     *slog.Logger
	timerWheel *algorithm.TimingWheel
}

// NewSchedulerManager creates a new SchedulerManager instance.
func NewSchedulerManager(repo domain.SchedulerRepository, logger *slog.Logger) *SchedulerManager {
	tw, err := algorithm.NewTimingWheel(time.Second, 3600)
	if err != nil {
		// 初始化失败属于严重配置错误
		logger.Error("failed to create timing wheel", "error", err)
		return nil
	}

	manager := &SchedulerManager{
		repo:       repo,
		logger:     logger,
		timerWheel: tw, // 1s 刻度，一小时周期
	}
	// 启动时间轮
	manager.timerWheel.Start()
	return manager
}

// ScheduleDelayJob 调度一个延迟任务。
// 与传统的 time.After 不同，时间轮可以在极低资源消耗下管理百万级别的延迟任务。
func (m *SchedulerManager) ScheduleDelayJob(ctx context.Context, delay time.Duration, jobID uint64) {
	m.logger.InfoContext(ctx, "scheduling delay job", "job_id", jobID, "delay", delay)

	m.timerWheel.AddTask(delay, func() {
		// 时间轮触发后的执行逻辑
		// 注意：此处在独立 goroutine 中运行
		innerCtx := context.Background()
		if err := m.RunJob(innerCtx, jobID); err != nil {
			m.logger.ErrorContext(innerCtx, "failed to run delay job from timer wheel", "job_id", jobID, "error", err)
		}
	})
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
	m.logger.InfoContext(ctx, "job started execution", "job_id", id)

	go func() {
		// 模拟实际任务执行逻辑
		time.Sleep(100 * time.Millisecond)

		endTime := time.Now()
		log.EndTime = &endTime
		log.Duration = endTime.Sub(log.StartTime).Milliseconds()
		log.Status = "SUCCESS"
		log.Result = "Executed successfully"

		if err := m.repo.SaveJobLog(context.Background(), log); err != nil {
			m.logger.Error("failed to save job log after execution", "job_id", id, "error", err)
		}

		job.Status = domain.JobStatusEnabled
		if err := m.repo.SaveJob(context.Background(), job); err != nil {
			m.logger.Error("failed to reset job status after execution", "job_id", id, "error", err)
		}
		m.logger.Info("job execution completed", "job_id", id, "status", log.Status)
	}()

	return nil
}
