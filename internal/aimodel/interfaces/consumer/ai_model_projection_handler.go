// 生成摘要：新增AI模型事件消费处理器，用于驱动读模型投影更新。
package consumer

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/segmentio/kafka-go"
	"github.com/wyfcoding/ecommerce/internal/aimodel/application"
	"github.com/wyfcoding/ecommerce/internal/aimodel/domain"
)

// AIModelProjectionHandler 处理AI模型事件并更新读模型。
type AIModelProjectionHandler struct {
	projector *application.AIModelProjectionService
	logger    *slog.Logger
}

// NewAIModelProjectionHandler 创建事件消费处理器。
func NewAIModelProjectionHandler(projector *application.AIModelProjectionService, logger *slog.Logger) *AIModelProjectionHandler {
	return &AIModelProjectionHandler{
		projector: projector,
		logger:    logger,
	}
}

// Handle 处理 Kafka 消息并触发读模型投影。
func (h *AIModelProjectionHandler) Handle(ctx context.Context, msg kafka.Message) error {
	switch msg.Topic {
	case domain.AIModelCreatedEventType:
		var event domain.AIModelCreatedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			h.logger.ErrorContext(ctx, "failed to unmarshal model created event", "error", err)
			return err
		}
		return h.projector.OnModelCreated(ctx, &event)
	case domain.AIModelStatusUpdatedEventType:
		var event domain.AIModelStatusUpdatedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			h.logger.ErrorContext(ctx, "failed to unmarshal model status updated event", "error", err)
			return err
		}
		return h.projector.OnModelStatusUpdated(ctx, &event)
	default:
		h.logger.WarnContext(ctx, "unknown aimodel event topic", "topic", msg.Topic)
		return nil
	}
}
