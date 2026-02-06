// 生成摘要：新增客服事件消费处理器，用于驱动读模型投影更新。
package consumer

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/segmentio/kafka-go"
	"github.com/wyfcoding/ecommerce/internal/support/application"
	"github.com/wyfcoding/ecommerce/internal/support/domain"
)

// SupportProjectionHandler 处理客服事件并更新读模型。
type SupportProjectionHandler struct {
	projector *application.SupportProjectionService
	logger    *slog.Logger
}

// NewSupportProjectionHandler 创建事件消费处理器。
func NewSupportProjectionHandler(projector *application.SupportProjectionService, logger *slog.Logger) *SupportProjectionHandler {
	return &SupportProjectionHandler{
		projector: projector,
		logger:    logger,
	}
}

// Handle 处理 Kafka 消息并触发读模型投影。
func (h *SupportProjectionHandler) Handle(ctx context.Context, msg kafka.Message) error {
	switch msg.Topic {
	case domain.TicketCreatedEventType:
		var event domain.TicketCreatedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			h.logger.ErrorContext(ctx, "failed to unmarshal ticket created event", "error", err)
			return err
		}
		return h.projector.OnTicketCreated(ctx, &event)
	case domain.TicketUpdatedEventType:
		var event domain.TicketUpdatedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			h.logger.ErrorContext(ctx, "failed to unmarshal ticket updated event", "error", err)
			return err
		}
		return h.projector.OnTicketUpdated(ctx, &event)
	case domain.TicketMessageCreatedEventType:
		var event domain.TicketMessageCreatedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			h.logger.ErrorContext(ctx, "failed to unmarshal ticket message created event", "error", err)
			return err
		}
		return h.projector.OnTicketMessageCreated(ctx, &event)
	case domain.ConversationCreatedEventType:
		var event domain.ConversationCreatedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			h.logger.ErrorContext(ctx, "failed to unmarshal conversation created event", "error", err)
			return err
		}
		return h.projector.OnConversationCreated(ctx, &event)
	case domain.ConversationMessageCreatedEventType:
		var event domain.ConversationMessageCreatedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			h.logger.ErrorContext(ctx, "failed to unmarshal conversation message created event", "error", err)
			return err
		}
		return h.projector.OnConversationMessageCreated(ctx, &event)
	default:
		h.logger.WarnContext(ctx, "unknown support event topic", "topic", msg.Topic)
		return nil
	}
}
