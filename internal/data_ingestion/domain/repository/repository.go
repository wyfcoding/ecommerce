package repository

import (
	"context"
	"github.com/wyfcoding/ecommerce/internal/data_ingestion/domain/entity"
)

// DataIngestionRepository 数据摄取仓储接口
type DataIngestionRepository interface {
	// Source methods
	SaveSource(ctx context.Context, source *entity.IngestionSource) error
	GetSource(ctx context.Context, id uint64) (*entity.IngestionSource, error)
	ListSources(ctx context.Context, offset, limit int) ([]*entity.IngestionSource, int64, error)
	UpdateSource(ctx context.Context, source *entity.IngestionSource) error
	DeleteSource(ctx context.Context, id uint64) error

	// Job methods
	SaveJob(ctx context.Context, job *entity.IngestionJob) error
	GetJob(ctx context.Context, id uint64) (*entity.IngestionJob, error)
	ListJobs(ctx context.Context, sourceID uint64, offset, limit int) ([]*entity.IngestionJob, int64, error)
	UpdateJob(ctx context.Context, job *entity.IngestionJob) error
}
