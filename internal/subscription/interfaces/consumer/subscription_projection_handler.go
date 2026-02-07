// 生成摘要：新增订阅事件消费处理器，用于驱动读模型投影更新。
package consumer

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/segmentio/kafka-go"
	"github.com/wyfcoding/ecommerce/internal/subscription/application"
	"github.com/wyfcoding/ecommerce/internal/subscription/domain"
)

// SubscriptionProjectionHandler 处理订阅事件并更新读模型。
type SubscriptionProjectionHandler struct {
	projector *application.SubscriptionProjectionService
	logger    *slog.Logger
}

// NewSubscriptionProjectionHandler 创建事件消费处理器。
func NewSubscriptionProjectionHandler(projector *application.SubscriptionProjectionService, logger *slog.Logger) *SubscriptionProjectionHandler {
	return &SubscriptionProjectionHandler{
		projector: projector,
		logger:    logger,
	}
}

// Handle 处理 Kafka 消息并触发读模型投影。
func (h *SubscriptionProjectionHandler) Handle(ctx context.Context, msg kafka.Message) error {
	switch msg.Topic {
	case domain.SubscriptionPlanCreatedEventType:
		var event domain.SubscriptionPlanCreatedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			h.logger.ErrorContext(ctx, "failed to unmarshal subscription plan created event", "error", err)
			return err
		}
		return h.projector.OnPlanCreated(ctx, &event)
	case domain.SubscriptionPlanUpdatedEventType:
		var event domain.SubscriptionPlanUpdatedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			h.logger.ErrorContext(ctx, "failed to unmarshal subscription plan updated event", "error", err)
			return err
		}
		return h.projector.OnPlanUpdated(ctx, &event)
	case domain.SubscriptionPlanDeletedEventType:
		var event domain.SubscriptionPlanDeletedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			h.logger.ErrorContext(ctx, "failed to unmarshal subscription plan deleted event", "error", err)
			return err
		}
		return h.projector.OnPlanDeleted(ctx, &event)
	case domain.SubscriptionCreatedEventType:
		var event domain.SubscriptionCreatedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			h.logger.ErrorContext(ctx, "failed to unmarshal subscription created event", "error", err)
			return err
		}
		return h.projector.OnSubscriptionCreated(ctx, &event)
	case domain.SubscriptionCancelledEventType:
		var event domain.SubscriptionCancelledEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			h.logger.ErrorContext(ctx, "failed to unmarshal subscription cancelled event", "error", err)
			return err
		}
		return h.projector.OnSubscriptionCancelled(ctx, &event)
	case domain.SubscriptionRenewedEventType:
		var event domain.SubscriptionRenewedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			h.logger.ErrorContext(ctx, "failed to unmarshal subscription renewed event", "error", err)
			return err
		}
		return h.projector.OnSubscriptionRenewed(ctx, &event)
	default:
		h.logger.WarnContext(ctx, "unknown subscription event topic", "topic", msg.Topic)
		return nil
	}
}
