// 生成摘要：新增评论事件消费处理器，用于驱动读模型投影更新。
// 假设：Kafka topic 与事件类型一一对应，消息体为 JSON。
package consumer

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/segmentio/kafka-go"
	"github.com/wyfcoding/ecommerce/internal/review/application"
	"github.com/wyfcoding/ecommerce/internal/review/domain"
)

// ReviewProjectionHandler 处理评论事件并更新读模型。
type ReviewProjectionHandler struct {
	projector *application.ReviewProjectionService
	logger    *slog.Logger
}

// NewReviewProjectionHandler 创建事件消费处理器。
func NewReviewProjectionHandler(projector *application.ReviewProjectionService, logger *slog.Logger) *ReviewProjectionHandler {
	return &ReviewProjectionHandler{
		projector: projector,
		logger:    logger,
	}
}

// Handle 处理 Kafka 消息并触发读模型投影。
func (h *ReviewProjectionHandler) Handle(ctx context.Context, msg kafka.Message) error {
	switch msg.Topic {
	case domain.ReviewCreatedEventType:
		var event domain.ReviewCreatedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			h.logger.ErrorContext(ctx, "failed to unmarshal review created event", "error", err)
			return err
		}
		return h.projector.OnReviewCreated(ctx, &event)
	case domain.ReviewUpdatedEventType:
		var event domain.ReviewUpdatedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			h.logger.ErrorContext(ctx, "failed to unmarshal review updated event", "error", err)
			return err
		}
		return h.projector.OnReviewUpdated(ctx, &event)
	case domain.ReviewDeletedEventType:
		var event domain.ReviewDeletedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			h.logger.ErrorContext(ctx, "failed to unmarshal review deleted event", "error", err)
			return err
		}
		return h.projector.OnReviewDeleted(ctx, &event)
	default:
		h.logger.WarnContext(ctx, "unknown review event topic", "topic", msg.Topic)
		return nil
	}
}
