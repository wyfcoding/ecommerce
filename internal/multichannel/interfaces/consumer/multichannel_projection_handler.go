// 生成摘要：新增多渠道事件消费处理器，用于驱动读模型投影更新。
// 假设：Kafka topic 与事件类型一一对应，消息体为 JSON。
package consumer

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/segmentio/kafka-go"
	"github.com/wyfcoding/ecommerce/internal/multichannel/application"
	"github.com/wyfcoding/ecommerce/internal/multichannel/domain"
)

// MultiChannelProjectionHandler 处理多渠道事件并更新读模型。
type MultiChannelProjectionHandler struct {
	projector *application.MultiChannelProjectionService
	logger    *slog.Logger
}

// NewMultiChannelProjectionHandler 创建事件消费处理器。
func NewMultiChannelProjectionHandler(projector *application.MultiChannelProjectionService, logger *slog.Logger) *MultiChannelProjectionHandler {
	return &MultiChannelProjectionHandler{
		projector: projector,
		logger:    logger,
	}
}

// Handle 处理 Kafka 消息并触发读模型投影。
func (h *MultiChannelProjectionHandler) Handle(ctx context.Context, msg kafka.Message) error {
	switch msg.Topic {
	case domain.ChannelRegisteredEventType:
		var event domain.ChannelRegisteredEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			h.logger.ErrorContext(ctx, "failed to unmarshal channel registered event", "error", err)
			return err
		}
		return h.projector.OnChannelRegistered(ctx, &event)
	case domain.ChannelOrderCreatedEventType:
		var event domain.ChannelOrderCreatedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			h.logger.ErrorContext(ctx, "failed to unmarshal channel order created event", "error", err)
			return err
		}
		return h.projector.OnChannelOrderCreated(ctx, &event)
	default:
		h.logger.WarnContext(ctx, "unknown multichannel event topic", "topic", msg.Topic)
		return nil
	}
}
