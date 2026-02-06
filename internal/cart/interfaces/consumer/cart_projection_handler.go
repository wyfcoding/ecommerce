// 生成摘要：新增购物车事件消费处理器，用于驱动读模型投影更新。
// 假设：Kafka topic 与事件类型一一对应，消息体为 JSON。
package consumer

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/segmentio/kafka-go"
	"github.com/wyfcoding/ecommerce/internal/cart/application"
	"github.com/wyfcoding/ecommerce/internal/cart/domain"
)

// CartProjectionHandler 处理购物车事件并更新读模型。
type CartProjectionHandler struct {
	projector *application.CartProjectionService
	logger    *slog.Logger
}

// NewCartProjectionHandler 创建事件消费处理器。
func NewCartProjectionHandler(projector *application.CartProjectionService, logger *slog.Logger) *CartProjectionHandler {
	return &CartProjectionHandler{
		projector: projector,
		logger:    logger,
	}
}

// Handle 处理 Kafka 消息并触发读模型投影。
func (h *CartProjectionHandler) Handle(ctx context.Context, msg kafka.Message) error {
	switch msg.Topic {
	case domain.CartItemAddedEventType:
		var event domain.CartItemAddedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			h.logger.ErrorContext(ctx, "failed to unmarshal cart item added event", "error", err)
			return err
		}
		return h.projector.OnItemAdded(ctx, &event)
	case domain.CartItemUpdatedEventType:
		var event domain.CartItemUpdatedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			h.logger.ErrorContext(ctx, "failed to unmarshal cart item updated event", "error", err)
			return err
		}
		return h.projector.OnItemUpdated(ctx, &event)
	case domain.CartItemRemovedEventType:
		var event domain.CartItemRemovedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			h.logger.ErrorContext(ctx, "failed to unmarshal cart item removed event", "error", err)
			return err
		}
		return h.projector.OnItemRemoved(ctx, &event)
	case domain.CartClearedEventType:
		var event domain.CartClearedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			h.logger.ErrorContext(ctx, "failed to unmarshal cart cleared event", "error", err)
			return err
		}
		return h.projector.OnCleared(ctx, &event)
	case domain.CartMergedEventType:
		var event domain.CartsMergedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			h.logger.ErrorContext(ctx, "failed to unmarshal cart merged event", "error", err)
			return err
		}
		return h.projector.OnMerged(ctx, &event)
	case domain.CartCouponAppliedEventType:
		var event domain.CouponAppliedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			h.logger.ErrorContext(ctx, "failed to unmarshal cart coupon applied event", "error", err)
			return err
		}
		return h.projector.OnCouponApplied(ctx, &event)
	case domain.CartCouponRemovedEventType:
		var event domain.CouponRemovedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			h.logger.ErrorContext(ctx, "failed to unmarshal cart coupon removed event", "error", err)
			return err
		}
		return h.projector.OnCouponRemoved(ctx, &event)
	default:
		h.logger.WarnContext(ctx, "unknown cart event topic", "topic", msg.Topic)
		return nil
	}
}
