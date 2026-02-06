// 生成摘要：新增推荐事件消费处理器，用于驱动读模型投影更新。
// 假设：Kafka topic 与事件类型一一对应，消息体为 JSON。
package consumer

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/segmentio/kafka-go"
	"github.com/wyfcoding/ecommerce/internal/recommendation/application"
	"github.com/wyfcoding/ecommerce/internal/recommendation/domain"
)

// RecommendationProjectionHandler 处理推荐事件并更新读模型。
type RecommendationProjectionHandler struct {
	projector *application.RecommendationProjectionService
	logger    *slog.Logger
}

// NewRecommendationProjectionHandler 创建事件消费处理器。
func NewRecommendationProjectionHandler(projector *application.RecommendationProjectionService, logger *slog.Logger) *RecommendationProjectionHandler {
	return &RecommendationProjectionHandler{
		projector: projector,
		logger:    logger,
	}
}

// Handle 处理 Kafka 消息并触发读模型投影。
func (h *RecommendationProjectionHandler) Handle(ctx context.Context, msg kafka.Message) error {
	switch msg.Topic {
	case domain.RecommendationChangedEventType:
		var event domain.RecommendationChangedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			h.logger.ErrorContext(ctx, "failed to unmarshal recommendation changed event", "error", err)
			return err
		}
		return h.projector.OnRecommendationChanged(ctx, &event)
	case domain.RecommendationDeletedEventType:
		var event domain.RecommendationDeletedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			h.logger.ErrorContext(ctx, "failed to unmarshal recommendation deleted event", "error", err)
			return err
		}
		return h.projector.OnRecommendationDeleted(ctx, &event)
	default:
		h.logger.WarnContext(ctx, "unknown recommendation event topic", "topic", msg.Topic)
		return nil
	}
}
