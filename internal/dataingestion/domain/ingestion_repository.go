package domain

import (
	"context"
)

// DataIngestionRepository 是数据摄取模块的仓储接口。
type DataIngestionRepository interface {
	// --- Source methods ---
	SaveSource(ctx context.Context, source *IngestionSource) error
	GetSource(ctx context.Context, id uint64) (*IngestionSource, error)
	ListSources(ctx context.Context, offset, limit int) ([]*IngestionSource, int64, error)
	UpdateSource(ctx context.Context, source *IngestionSource) error
	DeleteSource(ctx context.Context, id uint64) error

	// --- Job methods ---
	SaveJob(ctx context.Context, job *IngestionJob) error
	GetJob(ctx context.Context, id uint64) (*IngestionJob, error)
	ListJobs(ctx context.Context, sourceID uint64, offset, limit int) ([]*IngestionJob, int64, error)
	UpdateJob(ctx context.Context, job *IngestionJob) error

	// --- Event methods ---
	SaveEvent(ctx context.Context, event *IngestedEvent) error
	SaveEvents(ctx context.Context, events []*IngestedEvent) error
}
