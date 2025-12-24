package application

import (
	"context"
	"time"

	"github.com/wyfcoding/ecommerce/internal/dataingestion/domain"
)

// DataIngestionService 作为数据采集操作的门面。
type DataIngestionService struct {
	manager *DataIngestionManager
	query   *DataIngestionQuery
}

// NewDataIngestionService 创建数据采集服务门面实例。
func NewDataIngestionService(manager *DataIngestionManager, query *DataIngestionQuery) *DataIngestionService {
	return &DataIngestionService{
		manager: manager,
		query:   query,
	}
}

// --- 写操作（委托给 Manager）---

// RegisterSource 注册一个新的数据源。
func (s *DataIngestionService) RegisterSource(ctx context.Context, name string, sourceType domain.SourceType, config, description string) (*domain.IngestionSource, error) {
	return s.manager.RegisterSource(ctx, name, sourceType, config, description)
}

// TriggerIngestion 手动触发指定数据源的采集任务。
func (s *DataIngestionService) TriggerIngestion(ctx context.Context, sourceID uint64) (*domain.IngestionJob, error) {
	return s.manager.TriggerIngestion(ctx, sourceID)
}

// IngestEvent 实时采集单条事件数据。
func (s *DataIngestionService) IngestEvent(ctx context.Context, eventType, eventData, source string, timestamp time.Time) error {
	return s.manager.IngestEvent(ctx, eventType, eventData, source, timestamp)
}

// IngestBatch 批量采集多个事件数据。
func (s *DataIngestionService) IngestBatch(ctx context.Context, events []*domain.IngestedEvent) error {
	return s.manager.IngestBatch(ctx, events)
}

// --- 读操作（委托给 Query）---

// ListSources 列出所有注册的数据源。
func (s *DataIngestionService) ListSources(ctx context.Context, page, pageSize int) ([]*domain.IngestionSource, int64, error) {
	return s.query.ListSources(ctx, page, pageSize)
}

// ListJobs 列出采集任务历史记录。
func (s *DataIngestionService) ListJobs(ctx context.Context, sourceID uint64, page, pageSize int) ([]*domain.IngestionJob, int64, error) {
	return s.query.ListJobs(ctx, sourceID, page, pageSize)
}
