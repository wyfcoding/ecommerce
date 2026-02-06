// 生成摘要：新增售后事件消费处理器，用于驱动读模型投影更新。
package consumer

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/segmentio/kafka-go"
	"github.com/wyfcoding/ecommerce/internal/aftersales/application"
	"github.com/wyfcoding/ecommerce/internal/aftersales/domain"
)

// AfterSalesProjectionHandler 处理售后事件并更新读模型。
type AfterSalesProjectionHandler struct {
	projector *application.AfterSalesProjectionService
	logger    *slog.Logger
}

// NewAfterSalesProjectionHandler 创建事件消费处理器。
func NewAfterSalesProjectionHandler(projector *application.AfterSalesProjectionService, logger *slog.Logger) *AfterSalesProjectionHandler {
	return &AfterSalesProjectionHandler{
		projector: projector,
		logger:    logger,
	}
}

// Handle 处理 Kafka 消息并触发读模型投影。
func (h *AfterSalesProjectionHandler) Handle(ctx context.Context, msg kafka.Message) error {
	switch msg.Topic {
	case domain.AfterSalesCreatedEventType:
		var event domain.AfterSalesCreatedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			h.logger.ErrorContext(ctx, "failed to unmarshal aftersales created event", "error", err)
			return err
		}
		return h.projector.OnAfterSalesCreated(ctx, &event)
	case domain.AfterSalesStatusUpdatedEventType:
		var event domain.AfterSalesStatusUpdatedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			h.logger.ErrorContext(ctx, "failed to unmarshal aftersales status updated event", "error", err)
			return err
		}
		return h.projector.OnAfterSalesStatusUpdated(ctx, &event)
	case domain.AfterSalesSupportTicketCreatedType:
		var event domain.SupportTicketCreatedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			h.logger.ErrorContext(ctx, "failed to unmarshal support ticket created event", "error", err)
			return err
		}
		return h.projector.OnSupportTicketCreated(ctx, &event)
	case domain.AfterSalesSupportTicketUpdatedType:
		var event domain.SupportTicketUpdatedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			h.logger.ErrorContext(ctx, "failed to unmarshal support ticket updated event", "error", err)
			return err
		}
		return h.projector.OnSupportTicketUpdated(ctx, &event)
	case domain.AfterSalesSupportTicketMessageType:
		var event domain.SupportTicketMessageCreatedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			h.logger.ErrorContext(ctx, "failed to unmarshal support ticket message event", "error", err)
			return err
		}
		return h.projector.OnSupportTicketMessageCreated(ctx, &event)
	case domain.AfterSalesConfigUpdatedEventType:
		var event domain.AfterSalesConfigUpdatedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			h.logger.ErrorContext(ctx, "failed to unmarshal aftersales config event", "error", err)
			return err
		}
		return h.projector.OnConfigUpdated(ctx, &event)
	default:
		h.logger.WarnContext(ctx, "unknown aftersales event topic", "topic", msg.Topic)
		return nil
	}
}
