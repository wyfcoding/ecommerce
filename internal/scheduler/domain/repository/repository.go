package repository

import (
	"context"
	"github.com/wyfcoding/ecommerce/internal/scheduler/domain/entity" // 导入调度领域的实体定义。
)

// SchedulerRepository 是调度模块的仓储接口。
// 它定义了对定时任务和任务日志实体进行数据持久化操作的契约。
// 仓储接口属于领域层，旨在将领域逻辑与数据存储的实现细节解耦。
type SchedulerRepository interface {
	// --- 任务管理 (Job methods) ---

	// SaveJob 将定时任务实体保存到数据存储中。
	// 如果任务已存在，则更新；如果不存在，则创建。
	// ctx: 上下文。
	// job: 待保存的定时任务实体。
	SaveJob(ctx context.Context, job *entity.Job) error
	// GetJob 根据ID获取定时任务实体。
	GetJob(ctx context.Context, id uint64) (*entity.Job, error)
	// GetJobByName 根据任务名称获取定时任务实体。
	GetJobByName(ctx context.Context, name string) (*entity.Job, error)
	// ListJobs 列出所有定时任务实体，支持通过状态过滤和分页。
	ListJobs(ctx context.Context, status *entity.JobStatus, offset, limit int) ([]*entity.Job, int64, error)
	// DeleteJob 根据ID删除定时任务实体。
	DeleteJob(ctx context.Context, id uint64) error

	// --- 日志管理 (JobLog methods) ---

	// SaveJobLog 将任务日志实体保存到数据存储中。
	SaveJobLog(ctx context.Context, log *entity.JobLog) error
	// GetJobLog 根据ID获取任务日志实体。
	GetJobLog(ctx context.Context, id uint64) (*entity.JobLog, error)
	// ListJobLogs 列出指定任务ID的所有任务日志实体，支持分页。
	ListJobLogs(ctx context.Context, jobID uint64, offset, limit int) ([]*entity.JobLog, int64, error)
}
