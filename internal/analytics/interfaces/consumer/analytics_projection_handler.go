// 生成摘要：新增分析事件消费处理器，用于驱动读模型投影更新。
package consumer

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/segmentio/kafka-go"
	"github.com/wyfcoding/ecommerce/internal/analytics/application"
	"github.com/wyfcoding/ecommerce/internal/analytics/domain"
)

// AnalyticsProjectionHandler 处理分析事件并更新读模型。
type AnalyticsProjectionHandler struct {
	projector *application.AnalyticsProjectionService
	logger    *slog.Logger
}

// NewAnalyticsProjectionHandler 创建事件消费处理器。
func NewAnalyticsProjectionHandler(projector *application.AnalyticsProjectionService, logger *slog.Logger) *AnalyticsProjectionHandler {
	return &AnalyticsProjectionHandler{
		projector: projector,
		logger:    logger,
	}
}

// Handle 处理 Kafka 消息并触发读模型投影。
func (h *AnalyticsProjectionHandler) Handle(ctx context.Context, msg kafka.Message) error {
	switch msg.Topic {
	case domain.MetricRecordedEventType:
		var event domain.MetricRecordedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			h.logger.ErrorContext(ctx, "failed to unmarshal metric recorded event", "error", err)
			return err
		}
		return h.projector.OnMetricRecorded(ctx, &event)
	case domain.MetricDeletedEventType:
		var event domain.MetricDeletedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			h.logger.ErrorContext(ctx, "failed to unmarshal metric deleted event", "error", err)
			return err
		}
		return h.projector.OnMetricDeleted(ctx, &event)
	case domain.DashboardCreatedEventType:
		var event domain.DashboardCreatedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			h.logger.ErrorContext(ctx, "failed to unmarshal dashboard created event", "error", err)
			return err
		}
		return h.projector.OnDashboardCreated(ctx, &event)
	case domain.DashboardUpdatedEventType:
		var event domain.DashboardUpdatedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			h.logger.ErrorContext(ctx, "failed to unmarshal dashboard updated event", "error", err)
			return err
		}
		return h.projector.OnDashboardUpdated(ctx, &event)
	case domain.DashboardDeletedEventType:
		var event domain.DashboardDeletedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			h.logger.ErrorContext(ctx, "failed to unmarshal dashboard deleted event", "error", err)
			return err
		}
		return h.projector.OnDashboardDeleted(ctx, &event)
	case domain.ReportCreatedEventType:
		var event domain.ReportCreatedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			h.logger.ErrorContext(ctx, "failed to unmarshal report created event", "error", err)
			return err
		}
		return h.projector.OnReportCreated(ctx, &event)
	case domain.ReportUpdatedEventType:
		var event domain.ReportUpdatedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			h.logger.ErrorContext(ctx, "failed to unmarshal report updated event", "error", err)
			return err
		}
		return h.projector.OnReportUpdated(ctx, &event)
	case domain.ReportPublishedEventType:
		var event domain.ReportPublishedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			h.logger.ErrorContext(ctx, "failed to unmarshal report published event", "error", err)
			return err
		}
		return h.projector.OnReportPublished(ctx, &event)
	case domain.ReportDeletedEventType:
		var event domain.ReportDeletedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			h.logger.ErrorContext(ctx, "failed to unmarshal report deleted event", "error", err)
			return err
		}
		return h.projector.OnReportDeleted(ctx, &event)
	default:
		h.logger.WarnContext(ctx, "unknown analytics event topic", "topic", msg.Topic)
		return nil
	}
}
