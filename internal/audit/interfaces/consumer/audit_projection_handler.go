// 生成摘要：新增审计事件消费处理器，用于驱动读模型投影更新。
package consumer

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/segmentio/kafka-go"
	"github.com/wyfcoding/ecommerce/internal/audit/application"
	"github.com/wyfcoding/ecommerce/internal/audit/domain"
)

// AuditProjectionHandler 处理审计事件并更新读模型。
type AuditProjectionHandler struct {
	projector *application.AuditProjectionService
	logger    *slog.Logger
}

// NewAuditProjectionHandler 创建事件消费处理器。
func NewAuditProjectionHandler(projector *application.AuditProjectionService, logger *slog.Logger) *AuditProjectionHandler {
	return &AuditProjectionHandler{
		projector: projector,
		logger:    logger,
	}
}

// Handle 处理 Kafka 消息并触发读模型投影。
func (h *AuditProjectionHandler) Handle(ctx context.Context, msg kafka.Message) error {
	switch msg.Topic {
	case domain.AuditLogCreatedEventType:
		var event domain.AuditLogCreatedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			h.logger.ErrorContext(ctx, "failed to unmarshal audit log created event", "error", err)
			return err
		}
		return h.projector.OnLogCreated(ctx, &event)
	case domain.AuditLogDeletedEventType:
		var event domain.AuditLogDeletedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			h.logger.ErrorContext(ctx, "failed to unmarshal audit log deleted event", "error", err)
			return err
		}
		return h.projector.OnLogDeleted(ctx, &event)
	case domain.AuditPolicyCreatedEventType:
		var event domain.AuditPolicyCreatedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			h.logger.ErrorContext(ctx, "failed to unmarshal audit policy created event", "error", err)
			return err
		}
		return h.projector.OnPolicyCreated(ctx, &event)
	case domain.AuditPolicyUpdatedEventType:
		var event domain.AuditPolicyUpdatedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			h.logger.ErrorContext(ctx, "failed to unmarshal audit policy updated event", "error", err)
			return err
		}
		return h.projector.OnPolicyUpdated(ctx, &event)
	case domain.AuditPolicyDeletedEventType:
		var event domain.AuditPolicyDeletedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			h.logger.ErrorContext(ctx, "failed to unmarshal audit policy deleted event", "error", err)
			return err
		}
		return h.projector.OnPolicyDeleted(ctx, &event)
	case domain.AuditReportCreatedEventType:
		var event domain.AuditReportCreatedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			h.logger.ErrorContext(ctx, "failed to unmarshal audit report created event", "error", err)
			return err
		}
		return h.projector.OnReportCreated(ctx, &event)
	case domain.AuditReportUpdatedEventType:
		var event domain.AuditReportUpdatedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			h.logger.ErrorContext(ctx, "failed to unmarshal audit report updated event", "error", err)
			return err
		}
		return h.projector.OnReportUpdated(ctx, &event)
	case domain.AuditReportGeneratedEventType:
		var event domain.AuditReportGeneratedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			h.logger.ErrorContext(ctx, "failed to unmarshal audit report generated event", "error", err)
			return err
		}
		return h.projector.OnReportGenerated(ctx, &event)
	case domain.AuditReportPublishedEventType:
		var event domain.AuditReportPublishedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			h.logger.ErrorContext(ctx, "failed to unmarshal audit report published event", "error", err)
			return err
		}
		return h.projector.OnReportPublished(ctx, &event)
	case domain.AuditReportDeletedEventType:
		var event domain.AuditReportDeletedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			h.logger.ErrorContext(ctx, "failed to unmarshal audit report deleted event", "error", err)
			return err
		}
		return h.projector.OnReportDeleted(ctx, &event)
	default:
		h.logger.WarnContext(ctx, "unknown audit event topic", "topic", msg.Topic)
		return nil
	}
}
