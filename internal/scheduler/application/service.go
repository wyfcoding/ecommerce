package application

import (
	"context"
	"errors" // 导入标准错误处理库。
	"time"   // 导入时间库。

	"github.com/wyfcoding/ecommerce/internal/scheduler/domain/entity"     // 导入调度领域的实体定义。
	"github.com/wyfcoding/ecommerce/internal/scheduler/domain/repository" // 导入调度领域的仓储接口。

	"log/slog" // 导入结构化日志库。
)

// SchedulerService 结构体定义了任务调度相关的应用服务。
// 它协调领域层和基础设施层，处理定时任务的创建、管理、执行和日志记录等业务逻辑。
type SchedulerService struct {
	repo   repository.SchedulerRepository // 依赖SchedulerRepository接口，用于数据持久化操作。
	logger *slog.Logger                   // 日志记录器，用于记录服务运行时的信息和错误。
}

// NewSchedulerService 创建并返回一个新的 SchedulerService 实例。
func NewSchedulerService(repo repository.SchedulerRepository, logger *slog.Logger) *SchedulerService {
	return &SchedulerService{
		repo:   repo,
		logger: logger,
	}
}

// CreateJob 创建一个新的定时任务。
// ctx: 上下文。
// name: 任务名称，必须唯一。
// desc: 任务描述。
// cron: Cron表达式，定义任务的执行时间。
// handler: 任务处理器的标识。
// params: 任务执行所需的参数（JSON字符串）。
// 返回创建成功的Job实体和可能发生的错误。
func (s *SchedulerService) CreateJob(ctx context.Context, name, desc, cron, handler, params string) (*entity.Job, error) {
	// 1. 检查任务名称是否已存在，确保任务名称的唯一性。
	existing, err := s.repo.GetJobByName(ctx, name)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to check existing job name", "job_name", name, "error", err)
		return nil, err
	}
	if existing != nil {
		return nil, errors.New("job name already exists")
	}

	// 2. 创建Job实体。
	job := &entity.Job{
		Name:        name,
		Description: desc,
		CronExpr:    cron,
		Handler:     handler,
		Params:      params,
		Status:      entity.JobStatusEnabled, // 新任务默认为启用状态。
	}

	// 3. 通过仓储接口保存任务。
	if err := s.repo.SaveJob(ctx, job); err != nil {
		s.logger.ErrorContext(ctx, "failed to save job", "job_name", name, "error", err)
		return nil, err
	}
	s.logger.InfoContext(ctx, "job created successfully", "job_id", job.ID, "job_name", name)
	return job, nil
}

// UpdateJob 更新指定ID的定时任务信息（Cron表达式和参数）。
// ctx: 上下文。
// id: 任务ID。
// cron: 新的Cron表达式。
// params: 新的参数（JSON字符串）。
// 返回可能发生的错误。
func (s *SchedulerService) UpdateJob(ctx context.Context, id uint64, cron, params string) error {
	job, err := s.repo.GetJob(ctx, id)
	if err != nil {
		return err
	}
	if job == nil {
		return errors.New("job not found")
	}

	// 更新任务的Cron表达式和参数。
	job.CronExpr = cron
	job.Params = params
	// 通过仓储接口保存更新后的任务。
	if err := s.repo.SaveJob(ctx, job); err != nil {
		s.logger.ErrorContext(ctx, "failed to update job", "job_id", id, "error", err)
		return err
	}
	s.logger.InfoContext(ctx, "job updated successfully", "job_id", id)
	return nil
}

// ToggleJobStatus 切换指定ID的定时任务的启用/禁用状态。
// ctx: 上下文。
// id: 任务ID。
// enable: 布尔值，true表示启用，false表示禁用。
// 返回可能发生的错误。
func (s *SchedulerService) ToggleJobStatus(ctx context.Context, id uint64, enable bool) error {
	job, err := s.repo.GetJob(ctx, id)
	if err != nil {
		return err
	}
	if job == nil {
		return errors.New("job not found")
	}

	// 根据enable参数设置任务状态。
	if enable {
		job.Status = entity.JobStatusEnabled
	} else {
		job.Status = entity.JobStatusDisabled
	}
	// 通过仓储接口保存更新后的任务。
	if err := s.repo.SaveJob(ctx, job); err != nil {
		s.logger.ErrorContext(ctx, "failed to toggle job status", "job_id", id, "enable", enable, "error", err)
		return err
	}
	s.logger.InfoContext(ctx, "job status toggled successfully", "job_id", id, "enable", enable)
	return nil
}

// RunJob 立即运行指定ID的定时任务。
// ctx: 上下文。
// id: 任务ID。
// 返回可能发生的错误。
func (s *SchedulerService) RunJob(ctx context.Context, id uint64) error {
	job, err := s.repo.GetJob(ctx, id)
	if err != nil {
		return err
	}
	if job == nil {
		return errors.New("job not found")
	}

	// 检查任务是否已经在运行中。
	if job.Status == entity.JobStatusRunning {
		return errors.New("job is already running")
	}

	// 1. 创建任务日志记录。
	log := &entity.JobLog{
		JobID:     uint64(job.ID),
		JobName:   job.Name,
		Handler:   job.Handler,
		Params:    job.Params,
		Status:    "RUNNING", // 日志状态为运行中。
		StartTime: time.Now(),
	}
	if err := s.repo.SaveJobLog(ctx, log); err != nil {
		s.logger.ErrorContext(ctx, "failed to save job log", "job_id", id, "error", err)
		return err
	}

	// 2. 更新任务状态为运行中，并记录最后运行时间、运行次数。
	job.Status = entity.JobStatusRunning
	now := time.Now()
	job.LastRunTime = &now
	job.RunCount++
	if err := s.repo.SaveJob(ctx, job); err != nil {
		s.logger.ErrorContext(ctx, "failed to update job status to running", "job_id", id, "error", err)
		// TODO: 如果这里失败，需要更新 log 状态为失败。
		return err
	}
	s.logger.InfoContext(ctx, "job started execution (simulated)", "job_id", id)

	// 3. 模拟任务执行。
	// 在实际系统中，这里会异步地触发实际的任务处理器执行，例如通过消息队列或独立的执行器。
	// 这里使用goroutine模拟后台执行，并使用 context.Background() 确保后台任务不依赖于请求上下文。
	go func() {
		// 模拟任务工作耗时。
		time.Sleep(1 * time.Second)

		// 模拟任务完成后的处理。
		endTime := time.Now()
		log.EndTime = &endTime
		log.Duration = endTime.Sub(log.StartTime).Milliseconds() // 计算任务耗时。
		log.Status = "SUCCESS"                                   // 模拟成功状态。
		log.Result = "Executed successfully"

		// 异步更新任务日志。
		_ = s.repo.SaveJobLog(context.Background(), log)
		s.logger.Info("job execution completed (simulated)", "job_id", id, "status", log.Status)

		// 异步更新任务状态回启用。
		job.Status = entity.JobStatusEnabled
		_ = s.repo.SaveJob(context.Background(), job)
		s.logger.Info("job status reset to enabled (simulated)", "job_id", id)
	}()

	return nil
}

// ListJobs 获取定时任务列表。
// ctx: 上下文。
// status: 筛选任务状态。
// page, pageSize: 分页参数。
// 返回任务列表、总数和可能发生的错误。
func (s *SchedulerService) ListJobs(ctx context.Context, status *int, page, pageSize int) ([]*entity.Job, int64, error) {
	offset := (page - 1) * pageSize
	var st *entity.JobStatus
	if status != nil { // 如果提供了状态，则按状态过滤。
		s := entity.JobStatus(*status)
		st = &s
	}
	return s.repo.ListJobs(ctx, st, offset, pageSize)
}

// ListJobLogs 获取任务日志列表。
// ctx: 上下文。
// jobID: 筛选日志的任务ID。
// page, pageSize: 分页参数。
// 返回任务日志列表、总数和可能发生的错误。
func (s *SchedulerService) ListJobLogs(ctx context.Context, jobID uint64, page, pageSize int) ([]*entity.JobLog, int64, error) {
	offset := (page - 1) * pageSize
	return s.repo.ListJobLogs(ctx, jobID, offset, pageSize)
}
