package application

import (
	"context"

	"github.com/wyfcoding/ecommerce/internal/data_ingestion/domain"
)

// DataIngestionQuery 处理数据采集的读操作。
type DataIngestionQuery struct {
	repo domain.DataIngestionRepository
}

// NewDataIngestionQuery creates a new DataIngestionQuery instance.
func NewDataIngestionQuery(repo domain.DataIngestionRepository) *DataIngestionQuery {
	return &DataIngestionQuery{
		repo: repo,
	}
}

// ListSources 获取数据源列表。
func (q *DataIngestionQuery) ListSources(ctx context.Context, page, pageSize int) ([]*domain.IngestionSource, int64, error) {
	offset := (page - 1) * pageSize
	return q.repo.ListSources(ctx, offset, pageSize)
}

// ListJobs 获取数据摄取任务列表。
func (q *DataIngestionQuery) ListJobs(ctx context.Context, sourceID uint64, page, pageSize int) ([]*domain.IngestionJob, int64, error) {
	offset := (page - 1) * pageSize
	return q.repo.ListJobs(ctx, sourceID, offset, pageSize)
}
