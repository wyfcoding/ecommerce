// 生成摘要：新增内容审核事件消费处理器，用于驱动读模型投影更新。
package consumer

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/segmentio/kafka-go"
	"github.com/wyfcoding/ecommerce/internal/contentmoderation/application"
	"github.com/wyfcoding/ecommerce/internal/contentmoderation/domain"
)

// ModerationProjectionHandler 处理内容审核事件并更新读模型。
type ModerationProjectionHandler struct {
	projector *application.ModerationProjectionService
	logger    *slog.Logger
}

// NewModerationProjectionHandler 创建事件消费处理器。
func NewModerationProjectionHandler(projector *application.ModerationProjectionService, logger *slog.Logger) *ModerationProjectionHandler {
	return &ModerationProjectionHandler{
		projector: projector,
		logger:    logger,
	}
}

// Handle 处理 Kafka 消息并触发读模型投影。
func (h *ModerationProjectionHandler) Handle(ctx context.Context, msg kafka.Message) error {
	switch msg.Topic {
	case domain.ModerationRecordCreatedEventType:
		var event domain.ModerationRecordCreatedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			h.logger.ErrorContext(ctx, "failed to unmarshal moderation record created event", "error", err)
			return err
		}
		return h.projector.OnRecordCreated(ctx, &event)
	case domain.ModerationRecordUpdatedEventType:
		var event domain.ModerationRecordUpdatedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			h.logger.ErrorContext(ctx, "failed to unmarshal moderation record updated event", "error", err)
			return err
		}
		return h.projector.OnRecordUpdated(ctx, &event)
	case domain.ModerationRecordDeletedEventType:
		var event domain.ModerationRecordDeletedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			h.logger.ErrorContext(ctx, "failed to unmarshal moderation record deleted event", "error", err)
			return err
		}
		return h.projector.OnRecordDeleted(ctx, &event)
	case domain.SensitiveWordCreatedEventType:
		var event domain.SensitiveWordCreatedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			h.logger.ErrorContext(ctx, "failed to unmarshal sensitive word created event", "error", err)
			return err
		}
		return h.projector.OnSensitiveWordCreated(ctx, &event)
	case domain.SensitiveWordUpdatedEventType:
		var event domain.SensitiveWordUpdatedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			h.logger.ErrorContext(ctx, "failed to unmarshal sensitive word updated event", "error", err)
			return err
		}
		return h.projector.OnSensitiveWordUpdated(ctx, &event)
	case domain.SensitiveWordDeletedEventType:
		var event domain.SensitiveWordDeletedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			h.logger.ErrorContext(ctx, "failed to unmarshal sensitive word deleted event", "error", err)
			return err
		}
		return h.projector.OnSensitiveWordDeleted(ctx, &event)
	default:
		h.logger.WarnContext(ctx, "unknown moderation event topic", "topic", msg.Topic)
		return nil
	}
}
