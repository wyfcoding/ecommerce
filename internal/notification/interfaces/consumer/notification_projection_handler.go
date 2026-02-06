// 生成摘要：新增通知事件消费处理器，用于驱动读模型投影更新。
package consumer

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/segmentio/kafka-go"
	"github.com/wyfcoding/ecommerce/internal/notification/application"
	"github.com/wyfcoding/ecommerce/internal/notification/domain"
)

// NotificationProjectionHandler 处理通知事件并更新读模型。
type NotificationProjectionHandler struct {
	projector *application.NotificationProjectionService
	logger    *slog.Logger
}

// NewNotificationProjectionHandler 创建事件消费处理器。
func NewNotificationProjectionHandler(projector *application.NotificationProjectionService, logger *slog.Logger) *NotificationProjectionHandler {
	return &NotificationProjectionHandler{
		projector: projector,
		logger:    logger,
	}
}

// Handle 处理 Kafka 消息并触发读模型投影。
func (h *NotificationProjectionHandler) Handle(ctx context.Context, msg kafka.Message) error {
	switch msg.Topic {
	case domain.NotificationCreatedEventType:
		var event domain.NotificationCreatedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			h.logger.ErrorContext(ctx, "failed to unmarshal notification created event", "error", err)
			return err
		}
		return h.projector.OnNotificationCreated(ctx, &event)
	case domain.NotificationReadEventType:
		var event domain.NotificationReadEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			h.logger.ErrorContext(ctx, "failed to unmarshal notification read event", "error", err)
			return err
		}
		return h.projector.OnNotificationRead(ctx, &event)
	case domain.NotificationDeletedEventType:
		var event domain.NotificationDeletedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			h.logger.ErrorContext(ctx, "failed to unmarshal notification deleted event", "error", err)
			return err
		}
		return h.projector.OnNotificationDeleted(ctx, &event)
	case domain.NotificationTemplateCreatedEventType:
		var event domain.NotificationTemplateCreatedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			h.logger.ErrorContext(ctx, "failed to unmarshal notification template created event", "error", err)
			return err
		}
		return h.projector.OnTemplateCreated(ctx, &event)
	default:
		h.logger.WarnContext(ctx, "unknown notification event topic", "topic", msg.Topic)
		return nil
	}
}
