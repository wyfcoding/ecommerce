package repository

import (
	"context"
	"github.com/wyfcoding/ecommerce/internal/data_ingestion/domain/entity" // 导入数据摄取领域的实体定义。
)

// DataIngestionRepository 是数据摄取模块的仓储接口。
// 它定义了对摄取源和摄取任务实体进行数据持久化操作的契约。
// 仓储接口属于领域层，旨在将领域逻辑与数据存储的实现细节解耦。
type DataIngestionRepository interface {
	// --- Source methods ---

	// SaveSource 将数据源实体保存到数据存储中。
	// 如果数据源已存在，则更新；如果不存在，则创建。
	// ctx: 上下文。
	// source: 待保存的数据源实体。
	SaveSource(ctx context.Context, source *entity.IngestionSource) error
	// GetSource 根据ID获取数据源实体。
	GetSource(ctx context.Context, id uint64) (*entity.IngestionSource, error)
	// ListSources 列出所有数据源实体，支持分页。
	ListSources(ctx context.Context, offset, limit int) ([]*entity.IngestionSource, int64, error)
	// UpdateSource 更新数据源实体的信息。
	UpdateSource(ctx context.Context, source *entity.IngestionSource) error
	// DeleteSource 根据ID删除数据源实体。
	DeleteSource(ctx context.Context, id uint64) error

	// --- Job methods ---

	// SaveJob 将摄取任务实体保存到数据存储中。
	SaveJob(ctx context.Context, job *entity.IngestionJob) error
	// GetJob 根据ID获取摄取任务实体。
	GetJob(ctx context.Context, id uint64) (*entity.IngestionJob, error)
	// ListJobs 列出指定数据源ID的所有摄取任务实体，支持分页。
	ListJobs(ctx context.Context, sourceID uint64, offset, limit int) ([]*entity.IngestionJob, int64, error)
	// UpdateJob 更新摄取任务实体的信息。
	UpdateJob(ctx context.Context, job *entity.IngestionJob) error
}
