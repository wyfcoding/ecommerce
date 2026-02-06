package consumer

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/segmentio/kafka-go"
	"github.com/wyfcoding/ecommerce/internal/logistics/application"
	"github.com/wyfcoding/ecommerce/internal/logistics/domain"
)

// LogisticsProjectionHandler 处理物流事件并更新读模型。
type LogisticsProjectionHandler struct {
	projector *application.LogisticsProjectionService
	logger    *slog.Logger
}

// NewLogisticsProjectionHandler 创建事件消费处理器。
func NewLogisticsProjectionHandler(projector *application.LogisticsProjectionService, logger *slog.Logger) *LogisticsProjectionHandler {
	return &LogisticsProjectionHandler{projector: projector, logger: logger}
}

// Handle 处理 Kafka 消息并触发读模型投影。
func (h *LogisticsProjectionHandler) Handle(ctx context.Context, msg kafka.Message) error {
	switch msg.Topic {
	case domain.LogisticsCreatedEventType:
		var event domain.LogisticsCreatedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			h.logger.ErrorContext(ctx, "failed to unmarshal logistics created event", "error", err)
			return err
		}
		return h.projector.OnLogisticsCreated(ctx, &event)
	case domain.LogisticsStatusUpdatedEventType:
		var event domain.LogisticsStatusUpdatedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			h.logger.ErrorContext(ctx, "failed to unmarshal logistics status updated event", "error", err)
			return err
		}
		return h.projector.OnStatusUpdated(ctx, &event)
	case domain.LogisticsTraceAddedEventType:
		var event domain.LogisticsTraceAddedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			h.logger.ErrorContext(ctx, "failed to unmarshal logistics trace added event", "error", err)
			return err
		}
		return h.projector.OnTraceAdded(ctx, &event)
	case domain.LogisticsRiderAssignedEventType:
		var event domain.RiderAssignedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			h.logger.ErrorContext(ctx, "failed to unmarshal logistics rider assigned event", "error", err)
			return err
		}
		return h.projector.OnRiderAssigned(ctx, &event)
	default:
		h.logger.WarnContext(ctx, "unknown logistics event topic", "topic", msg.Topic)
		return nil
	}
}
