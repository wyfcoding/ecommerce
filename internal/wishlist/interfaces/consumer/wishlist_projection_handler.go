// 生成摘要：新增收藏夹事件消费处理器，用于驱动读模型投影更新。
// 假设：Kafka topic 与事件类型一一对应，消息体为 JSON。
package consumer

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/segmentio/kafka-go"
	"github.com/wyfcoding/ecommerce/internal/wishlist/application"
	"github.com/wyfcoding/ecommerce/internal/wishlist/domain"
)

// WishlistProjectionHandler 处理收藏夹事件并更新读模型。
type WishlistProjectionHandler struct {
	projector *application.WishlistProjectionService
	logger    *slog.Logger
}

// NewWishlistProjectionHandler 创建事件消费处理器。
func NewWishlistProjectionHandler(projector *application.WishlistProjectionService, logger *slog.Logger) *WishlistProjectionHandler {
	return &WishlistProjectionHandler{
		projector: projector,
		logger:    logger,
	}
}

// Handle 处理 Kafka 消息并触发读模型投影。
func (h *WishlistProjectionHandler) Handle(ctx context.Context, msg kafka.Message) error {
	switch msg.Topic {
	case domain.WishlistItemAddedEventType:
		var event domain.WishlistItemAddedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			h.logger.ErrorContext(ctx, "failed to unmarshal wishlist item added event", "error", err)
			return err
		}
		return h.projector.OnItemAdded(ctx, &event)
	case domain.WishlistItemRemovedEventType:
		var event domain.WishlistItemRemovedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			h.logger.ErrorContext(ctx, "failed to unmarshal wishlist item removed event", "error", err)
			return err
		}
		return h.projector.OnItemRemoved(ctx, &event)
	case domain.WishlistClearedEventType:
		var event domain.WishlistClearedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			h.logger.ErrorContext(ctx, "failed to unmarshal wishlist cleared event", "error", err)
			return err
		}
		return h.projector.OnCleared(ctx, &event)
	default:
		h.logger.WarnContext(ctx, "unknown wishlist event topic", "topic", msg.Topic)
		return nil
	}
}
