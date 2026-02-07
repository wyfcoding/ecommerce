package consumer

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/segmentio/kafka-go"
	"github.com/wyfcoding/ecommerce/internal/flashsale/application"
	"github.com/wyfcoding/ecommerce/internal/flashsale/domain"
)

// FlashSaleProjectionHandler 处理秒杀事件并更新读模型。
type FlashSaleProjectionHandler struct {
	projector *application.FlashSaleProjectionService
	logger    *slog.Logger
}

// NewFlashSaleProjectionHandler 创建事件消费处理器。
func NewFlashSaleProjectionHandler(projector *application.FlashSaleProjectionService, logger *slog.Logger) *FlashSaleProjectionHandler {
	return &FlashSaleProjectionHandler{
		projector: projector,
		logger:    logger,
	}
}

// Handle 处理 Kafka 消息并触发读模型投影。
func (h *FlashSaleProjectionHandler) Handle(ctx context.Context, msg kafka.Message) error {
	switch msg.Topic {
	case domain.FlashsaleCreatedEventType:
		var event domain.FlashSaleEventCreatedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			h.logger.ErrorContext(ctx, "failed to unmarshal flashsale created event", "error", err)
			return err
		}
		return h.projector.OnFlashsaleCreated(ctx, &event)
	case domain.FlashsaleOrderCreatedEventType:
		var event domain.FlashSaleOrderCreatedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			h.logger.ErrorContext(ctx, "failed to unmarshal flashsale order created event", "error", err)
			return err
		}
		return h.projector.OnOrderCreated(ctx, &event)
	case domain.FlashsaleOrderCancelledEventType:
		var event domain.FlashSaleOrderCancelledEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			h.logger.ErrorContext(ctx, "failed to unmarshal flashsale order cancelled event", "error", err)
			return err
		}
		return h.projector.OnOrderCancelled(ctx, &event)
	case domain.FlashsaleOrderPaidEventType:
		var event domain.FlashSaleOrderPaidEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			h.logger.ErrorContext(ctx, "failed to unmarshal flashsale order paid event", "error", err)
			return err
		}
		return h.projector.OnOrderPaid(ctx, &event)
	default:
		h.logger.WarnContext(ctx, "unknown flashsale event topic", "topic", msg.Topic)
		return nil
	}
}
