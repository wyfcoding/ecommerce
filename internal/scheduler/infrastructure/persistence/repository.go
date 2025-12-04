package persistence

import (
	"context"
	"errors" // 导入标准错误处理库。

	"github.com/wyfcoding/ecommerce/internal/scheduler/domain/entity"     // 导入调度领域的实体定义。
	"github.com/wyfcoding/ecommerce/internal/scheduler/domain/repository" // 导入调度领域的仓储接口。

	"gorm.io/gorm" // 导入GORM ORM框架。
)

type schedulerRepository struct {
	db *gorm.DB // GORM数据库连接实例。
}

// NewSchedulerRepository 创建并返回一个新的 schedulerRepository 实例。
func NewSchedulerRepository(db *gorm.DB) repository.SchedulerRepository {
	return &schedulerRepository{db: db}
}

// --- 任务管理 (Job methods) ---

// SaveJob 将定时任务实体保存到数据库。
// 如果实体已存在，则更新；如果不存在，则创建。
func (r *schedulerRepository) SaveJob(ctx context.Context, job *entity.Job) error {
	return r.db.WithContext(ctx).Save(job).Error
}

// GetJob 根据ID从数据库获取定时任务记录。
// 如果记录未找到，则返回nil。
func (r *schedulerRepository) GetJob(ctx context.Context, id uint64) (*entity.Job, error) {
	var job entity.Job
	if err := r.db.WithContext(ctx).First(&job, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // 如果记录未找到，返回nil。
		}
		return nil, err // 其他错误则返回。
	}
	return &job, nil
}

// GetJobByName 根据任务名称从数据库获取定时任务记录。
// 如果记录未找到，则返回nil。
func (r *schedulerRepository) GetJobByName(ctx context.Context, name string) (*entity.Job, error) {
	var job entity.Job
	// 按任务名称过滤。
	if err := r.db.WithContext(ctx).Where("name = ?", name).First(&job).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // 如果记录未找到，返回nil。
		}
		return nil, err // 其他错误则返回。
	}
	return &job, nil
}

// ListJobs 从数据库列出所有定时任务记录，支持通过状态过滤和分页。
func (r *schedulerRepository) ListJobs(ctx context.Context, status *entity.JobStatus, offset, limit int) ([]*entity.Job, int64, error) {
	var list []*entity.Job
	var total int64

	db := r.db.WithContext(ctx).Model(&entity.Job{})
	if status != nil { // 如果提供了状态，则按状态过滤。
		db = db.Where("status = ?", *status)
	}

	// 统计总记录数。
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 应用分页和排序（按ID降序）。
	if err := db.Offset(offset).Limit(limit).Order("id desc").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// DeleteJob 根据ID从数据库删除定时任务记录。
// GORM默认进行软删除。
func (r *schedulerRepository) DeleteJob(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&entity.Job{}, id).Error
}

// --- 日志管理 (JobLog methods) ---

// SaveJobLog 将任务日志实体保存到数据库。
func (r *schedulerRepository) SaveJobLog(ctx context.Context, log *entity.JobLog) error {
	return r.db.WithContext(ctx).Save(log).Error
}

// GetJobLog 根据ID从数据库获取任务日志记录。
// 如果记录未找到，则返回nil。
func (r *schedulerRepository) GetJobLog(ctx context.Context, id uint64) (*entity.JobLog, error) {
	var log entity.JobLog
	if err := r.db.WithContext(ctx).First(&log, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // 如果记录未找到，返回nil。
		}
		return nil, err // 其他错误则返回。
	}
	return &log, nil
}

// ListJobLogs 从数据库列出指定任务ID的所有任务日志记录，支持分页。
func (r *schedulerRepository) ListJobLogs(ctx context.Context, jobID uint64, offset, limit int) ([]*entity.JobLog, int64, error) {
	var list []*entity.JobLog
	var total int64

	db := r.db.WithContext(ctx).Model(&entity.JobLog{})
	if jobID > 0 { // 如果提供了任务ID，则按任务ID过滤。
		db = db.Where("job_id = ?", jobID)
	}

	// 统计总记录数。
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 应用分页和排序（按开始时间降序）。
	if err := db.Offset(offset).Limit(limit).Order("start_time desc").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}
