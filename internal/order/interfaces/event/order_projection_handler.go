// 生成摘要：新增订单事件消费处理器，用于驱动读模型投影更新。
// 假设：Kafka topic 与事件类型一一对应，消息体为 JSON。
package consumer

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/wyfcoding/ecommerce/internal/order/application"
	"github.com/wyfcoding/ecommerce/internal/order/domain"

	"github.com/segmentio/kafka-go"
)

// OrderProjectionHandler 处理订单事件并更新读模型。
type OrderProjectionHandler struct {
	projector *application.OrderProjectionService
	logger    *slog.Logger
}

// NewOrderProjectionHandler 创建事件消费处理器。
func NewOrderProjectionHandler(projector *application.OrderProjectionService, logger *slog.Logger) *OrderProjectionHandler {
	return &OrderProjectionHandler{
		projector: projector,
		logger:    logger,
	}
}

// Handle 处理 Kafka 消息并触发读模型投影。
func (h *OrderProjectionHandler) Handle(ctx context.Context, msg kafka.Message) error {
	switch msg.Topic {
	case "order.created":
		var event domain.OrderCreatedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			h.logger.ErrorContext(ctx, "failed to unmarshal order created event", "error", err)
			return err
		}
		return h.projector.OnOrderCreated(ctx, &event)
	case "order.paid":
		var event domain.OrderPaidEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			h.logger.ErrorContext(ctx, "failed to unmarshal order paid event", "error", err)
			return err
		}
		return h.projector.OnOrderPaid(ctx, &event)
	case "order.shipped":
		var event domain.OrderShippedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			h.logger.ErrorContext(ctx, "failed to unmarshal order shipped event", "error", err)
			return err
		}
		return h.projector.OnOrderShipped(ctx, &event)
	case "order.delivered":
		var event domain.OrderDeliveredEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			h.logger.ErrorContext(ctx, "failed to unmarshal order delivered event", "error", err)
			return err
		}
		return h.projector.OnOrderDelivered(ctx, &event)
	case "order.completed":
		var event domain.OrderCompletedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			h.logger.ErrorContext(ctx, "failed to unmarshal order completed event", "error", err)
			return err
		}
		return h.projector.OnOrderCompleted(ctx, &event)
	case "order.cancelled":
		var event domain.OrderCancelledEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			h.logger.ErrorContext(ctx, "failed to unmarshal order cancelled event", "error", err)
			return err
		}
		return h.projector.OnOrderCancelled(ctx, &event)
	case "order.confirmed":
		var event domain.OrderConfirmedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			h.logger.ErrorContext(ctx, "failed to unmarshal order confirmed event", "error", err)
			return err
		}
		return h.projector.OnOrderConfirmed(ctx, &event)
	default:
		h.logger.WarnContext(ctx, "unknown order event topic", "topic", msg.Topic)
		return nil
	}
}
