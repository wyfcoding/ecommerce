// 生成摘要：新增拼团事件消费处理器，用于驱动读模型投影更新。
// 假设：Kafka topic 与事件类型一一对应，消息体为 JSON。
package consumer

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/segmentio/kafka-go"
	"github.com/wyfcoding/ecommerce/internal/groupbuy/application"
	"github.com/wyfcoding/ecommerce/internal/groupbuy/domain"
)

// GroupbuyProjectionHandler 处理拼团事件并更新读模型。
type GroupbuyProjectionHandler struct {
	projector *application.GroupbuyProjectionService
	logger    *slog.Logger
}

// NewGroupbuyProjectionHandler 创建事件消费处理器。
func NewGroupbuyProjectionHandler(projector *application.GroupbuyProjectionService, logger *slog.Logger) *GroupbuyProjectionHandler {
	return &GroupbuyProjectionHandler{
		projector: projector,
		logger:    logger,
	}
}

// Handle 处理 Kafka 消息并触发读模型投影。
func (h *GroupbuyProjectionHandler) Handle(ctx context.Context, msg kafka.Message) error {
	switch msg.Topic {
	case domain.GroupbuyCreatedEventType:
		var event domain.GroupBuyCreatedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			h.logger.ErrorContext(ctx, "failed to unmarshal groupbuy created event", "error", err)
			return err
		}
		return h.projector.OnGroupbuyCreated(ctx, &event)
	case domain.GroupbuyJoinedEventType:
		var event domain.GroupBuyJoinedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			h.logger.ErrorContext(ctx, "failed to unmarshal groupbuy joined event", "error", err)
			return err
		}
		return h.projector.OnGroupbuyJoined(ctx, &event)
	case domain.GroupbuyCompletedEventType:
		var event domain.GroupBuyCompletedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			h.logger.ErrorContext(ctx, "failed to unmarshal groupbuy completed event", "error", err)
			return err
		}
		return h.projector.OnGroupbuyCompleted(ctx, &event)
	default:
		h.logger.WarnContext(ctx, "unknown groupbuy event topic", "topic", msg.Topic)
		return nil
	}
}
