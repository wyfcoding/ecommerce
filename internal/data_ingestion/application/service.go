package application

import (
	"context"
	"time"

	"github.com/wyfcoding/ecommerce/internal/data_ingestion/domain/entity"     // 导入数据摄取领域的实体定义。
	"github.com/wyfcoding/ecommerce/internal/data_ingestion/domain/repository" // 导入数据摄取领域的仓储接口。

	"log/slog" // 导入结构化日志库。
)

// DataIngestionService 结构体定义了数据摄取相关的应用服务。
// 它协调领域层和基础设施层，处理数据源的注册、摄取任务的触发和状态管理等业务逻辑。
type DataIngestionService struct {
	repo   repository.DataIngestionRepository // 依赖DataIngestionRepository接口，用于数据持久化操作。
	logger *slog.Logger                       // 日志记录器，用于记录服务运行时的信息和错误。
}

// NewDataIngestionService 创建并返回一个新的 DataIngestionService 实例。
func NewDataIngestionService(repo repository.DataIngestionRepository, logger *slog.Logger) *DataIngestionService {
	return &DataIngestionService{
		repo:   repo,
		logger: logger,
	}
}

// RegisterSource 注册一个新的数据源。
// ctx: 上下文。
// name: 数据源名称。
// sourceType: 数据源类型（例如，“Kafka”，“MySQL”）。
// config: 数据源的连接配置（例如，JSON字符串）。
// description: 数据源描述。
// 返回创建成功的IngestionSource实体和可能发生的错误。
func (s *DataIngestionService) RegisterSource(ctx context.Context, name string, sourceType entity.SourceType, config, description string) (*entity.IngestionSource, error) {
	source := entity.NewIngestionSource(name, sourceType, config, description) // 创建IngestionSource实体。
	// 通过仓储接口保存数据源。
	if err := s.repo.SaveSource(ctx, source); err != nil {
		s.logger.ErrorContext(ctx, "failed to save source", "name", name, "error", err)
		return nil, err
	}
	s.logger.InfoContext(ctx, "source registered successfully", "source_id", source.ID, "name", name)
	return source, nil
}

// ListSources 获取数据源列表。
// ctx: 上下文。
// page, pageSize: 分页参数。
// 返回数据源列表、总数和可能发生的错误。
func (s *DataIngestionService) ListSources(ctx context.Context, page, pageSize int) ([]*entity.IngestionSource, int64, error) {
	offset := (page - 1) * pageSize
	return s.repo.ListSources(ctx, offset, pageSize)
}

// TriggerIngestion 触发一个数据摄取任务。
// ctx: 上下文。
// sourceID: 待摄取数据的源ID。
// 返回新创建的IngestionJob实体和可能发生的错误。
func (s *DataIngestionService) TriggerIngestion(ctx context.Context, sourceID uint64) (*entity.IngestionJob, error) {
	// 获取数据源信息。
	source, err := s.repo.GetSource(ctx, sourceID)
	if err != nil {
		return nil, err
	}

	job := entity.NewIngestionJob(uint64(source.ID)) // 创建IngestionJob实体。
	// 通过仓储接口保存摄取任务。
	if err := s.repo.SaveJob(ctx, job); err != nil {
		s.logger.ErrorContext(ctx, "failed to save job", "source_id", sourceID, "error", err)
		return nil, err
	}
	s.logger.InfoContext(ctx, "ingestion job triggered", "job_id", job.ID, "source_id", sourceID)

	// 异步处理摄取任务。
	go s.processJob(job)

	return job, nil
}

// processJob 异步处理数据摄取任务的后台逻辑。
// job: 待处理的摄取任务实体。
func (s *DataIngestionService) processJob(job *entity.IngestionJob) {
	ctx := context.Background() // 使用一个新的背景上下文处理后台任务。
	job.Start()                 // 调用实体方法更新任务状态为运行中。
	s.repo.UpdateJob(ctx, job)  // 更新数据库中的任务状态。
	s.logger.InfoContext(ctx, "ingestion job started", "job_id", job.ID)

	// Simulate processing: 模拟数据处理过程。
	time.Sleep(2 * time.Second)

	// Success simulation: 模拟成功完成摄取。
	job.Complete(100)          // 调用实体方法更新任务状态为完成，并记录处理行数。
	s.repo.UpdateJob(ctx, job) // 更新数据库中的任务状态。
	s.logger.InfoContext(ctx, "ingestion job completed", "job_id", job.ID, "rows_processed", 100)

	// 更新数据源的最后运行时间。
	source, _ := s.repo.GetSource(ctx, job.SourceID)
	if source != nil {
		now := time.Now()
		source.LastRunAt = &now
		s.repo.UpdateSource(ctx, source)
	}
}

// ListJobs 获取数据摄取任务列表。
// ctx: 上下文。
// sourceID: 筛选任务的数据源ID。
// page, pageSize: 分页参数。
// 返回摄取任务列表、总数和可能发生的错误。
func (s *DataIngestionService) ListJobs(ctx context.Context, sourceID uint64, page, pageSize int) ([]*entity.IngestionJob, int64, error) {
	offset := (page - 1) * pageSize
	return s.repo.ListJobs(ctx, sourceID, offset, pageSize)
}

// IngestEvent 摄取单个事件。
// ctx: 上下文。
// eventType: 事件类型。
// eventData: 事件数据（JSON字符串）。
// source: 来源。
// timestamp: 事件时间戳。
// 返回可能发生的错误。
func (s *DataIngestionService) IngestEvent(ctx context.Context, eventType, eventData, source string, timestamp time.Time) error {
	event := entity.NewIngestedEvent(eventType, eventData, source, timestamp)
	if err := s.repo.SaveEvent(ctx, event); err != nil {
		s.logger.ErrorContext(ctx, "failed to save event", "event_type", eventType, "error", err)
		return err
	}
	s.logger.InfoContext(ctx, "event ingested successfully", "event_id", event.ID, "event_type", eventType)
	return nil
}

// IngestBatch 批量摄取事件。
// ctx: 上下文。
// events: 待摄取的事件列表。
// 返回可能发生的错误。
func (s *DataIngestionService) IngestBatch(ctx context.Context, events []*entity.IngestedEvent) error {
	if len(events) == 0 {
		return nil
	}
	if err := s.repo.SaveEvents(ctx, events); err != nil {
		s.logger.ErrorContext(ctx, "failed to save batch events", "count", len(events), "error", err)
		return err
	}
	s.logger.InfoContext(ctx, "batch events ingested successfully", "count", len(events))
	return nil
}
