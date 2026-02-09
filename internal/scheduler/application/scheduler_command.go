package application

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/wyfcoding/ecommerce/internal/scheduler/domain"
	algorithm "github.com/wyfcoding/pkg/algorithm/infra"
	"github.com/wyfcoding/pkg/lock"
	"github.com/wyfcoding/pkg/messagequeue"
	pkgScheduler "github.com/wyfcoding/pkg/scheduler"
)

// JobHandler 定义了任务处理函数的原型
type JobHandler func(ctx context.Context, params string) (string, error)

// SchedulerCommandService 处理调度任务和日志的写操作。
type SchedulerCommandService struct {
	repo       domain.SchedulerRepository
	publisher  messagequeue.EventPublisher
	logger     *slog.Logger
	timerWheel *algorithm.TimingWheel
	cronSched  *pkgScheduler.Scheduler
	distLock   lock.DistributedLock
	handlers   map[string]JobHandler
}

// NewSchedulerCommandService creates a new SchedulerCommandService instance.
func NewSchedulerCommandService(
	repo domain.SchedulerRepository,
	publisher messagequeue.EventPublisher,
	cronSched *pkgScheduler.Scheduler,
	distLock lock.DistributedLock,
	logger *slog.Logger,
) (*SchedulerCommandService, error) {
	tw, err := algorithm.NewTimingWheel(time.Second, 3600)
	if err != nil {
		return nil, err
	}

	service := &SchedulerCommandService{
		repo:       repo,
		publisher:  publisher,
		logger:     logger,
		timerWheel: tw,
		cronSched:  cronSched,
		distLock:   distLock,
		handlers:   make(map[string]JobHandler),
	}
	service.timerWheel.Start()
	return service, nil
}

// Start 启动调度器 (非阻塞启动，然后阻塞等待上下文取消)。
func (m *SchedulerCommandService) Start(ctx context.Context) error {
	m.cronSched.Start(ctx)
	// 阻塞直到上下文取消，模拟 Server 行为
	<-ctx.Done()
	return ctx.Err()
}

// Stop 停止调度器。
func (m *SchedulerCommandService) Stop(ctx context.Context) error {
	m.timerWheel.Stop()
	return m.cronSched.Stop(ctx)
}

// RegisterHandler 注册任务处理器
func (m *SchedulerCommandService) RegisterHandler(name string, handler JobHandler) {
	m.handlers[name] = handler
}

// LoadJobs 从数据库加载所有启用的 Cron 任务到调度器。
func (m *SchedulerCommandService) LoadJobs(ctx context.Context) error {
	status := domain.JobStatusEnabled
	jobs, _, err := m.repo.ListJobs(ctx, &status, 1, 1000) // 假设不超过 1000 个任务
	if err != nil {
		return err
	}

	for _, job := range jobs {
		if err := m.addJobToScheduler(ctx, job); err != nil {
			m.logger.ErrorContext(ctx, "failed to load job to scheduler", "job_name", job.Name, "error", err)
			continue
		}
	}
	return nil
}

// addJobToScheduler 辅助方法：将 domain.Job 转换为 pkg/scheduler 任务并添加。
func (m *SchedulerCommandService) addJobToScheduler(ctx context.Context, job *domain.Job) error {
	jobID := uint64(job.ID)
	cfg := pkgScheduler.JobConfig{
		Name:            job.Name,
		CronExpr:        job.CronExpr,
		Retry:           0,                // 数据库未存 retry 次数，默认 0
		Timeout:         30 * time.Minute, // 默认超时
		Lock:            m.distLock,       // 启用分布式锁
		LockTTL:         5 * time.Minute,
		AllowConcurrent: false,
	}

	return m.cronSched.AddJob(cfg, func(jobCtx context.Context) error {
		// 任务触发时调用 RunJob 逻辑
		return m.RunJob(jobCtx, jobID)
	})
}

// ScheduleDelayJob 调度一个延迟任务。
// 与传统的 time.After 不同，时间轮可以在极低资源消耗下管理百万级别的延迟任务。
func (m *SchedulerCommandService) ScheduleDelayJob(ctx context.Context, delay time.Duration, jobID uint64) {
	m.logger.InfoContext(ctx, "scheduling delay job", "job_id", jobID, "delay", delay)

	if err := m.timerWheel.AddTask(delay, func() {
		// 时间轮触发后的执行逻辑
		// 注意：此处在独立 goroutine 中运行
		innerCtx := context.Background()
		if err := m.RunJob(innerCtx, jobID); err != nil {
			m.logger.ErrorContext(innerCtx, "failed to run delay job from timer wheel", "job_id", jobID, "error", err)
		}
	}); err != nil {
		m.logger.ErrorContext(ctx, "failed to add task to timer wheel", "job_id", jobID, "error", err)
	}
}

// CreateJob 创建一个新的定时任务。
func (m *SchedulerCommandService) CreateJob(ctx context.Context, name, desc, cron, handler, params string) (*domain.Job, error) {
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

	// 添加到调度器
	if err := m.addJobToScheduler(ctx, job); err != nil {
		m.logger.WarnContext(ctx, "failed to add new job to scheduler immediately", "job_name", name, "error", err)
	}

	m.logger.InfoContext(ctx, "job created successfully", "job_id", job.ID, "job_name", name)
	return job, nil
}

// UpdateJob 更新指定ID的定时任务信息。
func (m *SchedulerCommandService) UpdateJob(ctx context.Context, id uint64, cron, params string) error {
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

	// 更新调度器
	if job.Status == domain.JobStatusEnabled {
		if err := m.addJobToScheduler(ctx, job); err != nil {
			m.logger.WarnContext(ctx, "failed to update job in scheduler", "job_id", id, "error", err)
		}
	}

	m.logger.InfoContext(ctx, "job updated successfully", "job_id", id)
	return nil
}

// ToggleJobStatus 切换定时任务的启用/禁用状态。
func (m *SchedulerCommandService) ToggleJobStatus(ctx context.Context, id uint64, enable bool) error {
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

	// 同步状态到调度器
	if enable {
		if err := m.addJobToScheduler(ctx, job); err != nil {
			m.logger.ErrorContext(ctx, "failed to enable job in scheduler", "job_id", id, "error", err)
		}
	} else {
		m.cronSched.RemoveJob(job.Name)
	}

	m.logger.InfoContext(ctx, "job status toggled successfully", "job_id", id, "enable", enable)
	return nil
}

// RunJob 立即运行指定ID的定时任务。
func (m *SchedulerCommandService) RunJob(ctx context.Context, id uint64) error {
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
		// 真实化执行：根据 Handler 名称查找并运行业务逻辑
		var (
			result string
			err    error
			status = "SUCCESS"
		)

		handler, ok := m.handlers[job.Handler]
		if !ok {
			err = fmt.Errorf("handler %s not found", job.Handler)
			status = "FAILED"
		} else {
			// 执行真实逻辑
			result, err = handler(context.Background(), job.Params)
			if err != nil {
				status = "FAILED"
			}
		}

		endTime := time.Now()
		log.EndTime = &endTime
		log.Duration = endTime.Sub(log.StartTime).Milliseconds()
		log.Status = status
		if err != nil {
			log.Result = err.Error()
		} else {
			log.Result = result
		}

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
