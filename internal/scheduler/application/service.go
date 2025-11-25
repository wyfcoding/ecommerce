package application

import (
	"context"
	"github.com/wyfcoding/ecommerce/internal/scheduler/domain/entity"
	"github.com/wyfcoding/ecommerce/internal/scheduler/domain/repository"
	"errors"
	"time"

	"log/slog"
)

type SchedulerService struct {
	repo   repository.SchedulerRepository
	logger *slog.Logger
}

func NewSchedulerService(repo repository.SchedulerRepository, logger *slog.Logger) *SchedulerService {
	return &SchedulerService{
		repo:   repo,
		logger: logger,
	}
}

// CreateJob 创建任务
func (s *SchedulerService) CreateJob(ctx context.Context, name, desc, cron, handler, params string) (*entity.Job, error) {
	// Check if name exists
	existing, _ := s.repo.GetJobByName(ctx, name)
	if existing != nil {
		return nil, errors.New("job name already exists")
	}

	job := &entity.Job{
		Name:        name,
		Description: desc,
		CronExpr:    cron,
		Handler:     handler,
		Params:      params,
		Status:      entity.JobStatusEnabled,
	}

	if err := s.repo.SaveJob(ctx, job); err != nil {
		return nil, err
	}
	return job, nil
}

// UpdateJob 更新任务
func (s *SchedulerService) UpdateJob(ctx context.Context, id uint64, cron, params string) error {
	job, err := s.repo.GetJob(ctx, id)
	if err != nil {
		return err
	}
	if job == nil {
		return errors.New("job not found")
	}

	job.CronExpr = cron
	job.Params = params
	return s.repo.SaveJob(ctx, job)
}

// ToggleJobStatus 切换任务状态
func (s *SchedulerService) ToggleJobStatus(ctx context.Context, id uint64, enable bool) error {
	job, err := s.repo.GetJob(ctx, id)
	if err != nil {
		return err
	}
	if job == nil {
		return errors.New("job not found")
	}

	if enable {
		job.Status = entity.JobStatusEnabled
	} else {
		job.Status = entity.JobStatusDisabled
	}
	return s.repo.SaveJob(ctx, job)
}

// RunJob 立即运行任务 (Mock execution)
func (s *SchedulerService) RunJob(ctx context.Context, id uint64) error {
	job, err := s.repo.GetJob(ctx, id)
	if err != nil {
		return err
	}
	if job == nil {
		return errors.New("job not found")
	}

	if job.Status == entity.JobStatusRunning {
		return errors.New("job is already running")
	}

	// Create Log
	log := &entity.JobLog{
		JobID:     uint64(job.ID),
		JobName:   job.Name,
		Handler:   job.Handler,
		Params:    job.Params,
		Status:    "RUNNING",
		StartTime: time.Now(),
	}
	if err := s.repo.SaveJobLog(ctx, log); err != nil {
		return err
	}

	// Update Job Status
	job.Status = entity.JobStatusRunning
	now := time.Now()
	job.LastRunTime = &now
	job.RunCount++
	if err := s.repo.SaveJob(ctx, job); err != nil {
		return err
	}

	// Mock Execution (Async in real world)
	go func() {
		// Simulate work
		time.Sleep(1 * time.Second)

		// Complete
		endTime := time.Now()
		log.EndTime = &endTime
		log.Duration = endTime.Sub(log.StartTime).Milliseconds()
		log.Status = "SUCCESS"
		log.Result = "Executed successfully"

		// Update Log
		// Note: In real app, use a new context or background context
		_ = s.repo.SaveJobLog(context.Background(), log)

		// Update Job
		job.Status = entity.JobStatusEnabled
		_ = s.repo.SaveJob(context.Background(), job)
	}()

	return nil
}

// ListJobs 任务列表
func (s *SchedulerService) ListJobs(ctx context.Context, status *int, page, pageSize int) ([]*entity.Job, int64, error) {
	offset := (page - 1) * pageSize
	var st *entity.JobStatus
	if status != nil {
		s := entity.JobStatus(*status)
		st = &s
	}
	return s.repo.ListJobs(ctx, st, offset, pageSize)
}

// ListJobLogs 日志列表
func (s *SchedulerService) ListJobLogs(ctx context.Context, jobID uint64, page, pageSize int) ([]*entity.JobLog, int64, error) {
	offset := (page - 1) * pageSize
	return s.repo.ListJobLogs(ctx, jobID, offset, pageSize)
}
