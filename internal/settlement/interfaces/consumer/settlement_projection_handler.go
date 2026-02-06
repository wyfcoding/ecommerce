// 生成摘要：新增结算事件消费处理器，用于驱动读模型投影更新。
// 假设：Kafka topic 与事件类型一一对应，消息体为 JSON。
package consumer

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/segmentio/kafka-go"
	"github.com/wyfcoding/ecommerce/internal/settlement/application"
	"github.com/wyfcoding/ecommerce/internal/settlement/domain"
)

// SettlementProjectionHandler 处理结算事件并更新读模型。
type SettlementProjectionHandler struct {
	projector *application.SettlementProjectionService
	logger    *slog.Logger
}

// NewSettlementProjectionHandler 创建事件消费处理器。
func NewSettlementProjectionHandler(projector *application.SettlementProjectionService, logger *slog.Logger) *SettlementProjectionHandler {
	return &SettlementProjectionHandler{
		projector: projector,
		logger:    logger,
	}
}

// Handle 处理 Kafka 消息并触发读模型投影。
func (h *SettlementProjectionHandler) Handle(ctx context.Context, msg kafka.Message) error {
	switch msg.Topic {
	case "settlement.created":
		var event domain.SettlementCreatedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			h.logger.ErrorContext(ctx, "failed to unmarshal settlement created event", "error", err)
			return err
		}
		return h.projector.OnSettlementCreated(ctx, &event)
	case "settlement.processed":
		var event domain.SettlementProcessedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			h.logger.ErrorContext(ctx, "failed to unmarshal settlement processed event", "error", err)
			return err
		}
		return h.projector.OnSettlementProcessed(ctx, &event)
	case "settlement.completed":
		var event domain.SettlementCompletedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			h.logger.ErrorContext(ctx, "failed to unmarshal settlement completed event", "error", err)
			return err
		}
		return h.projector.OnSettlementCompleted(ctx, &event)
	case "settlement.failed":
		var event domain.SettlementFailedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			h.logger.ErrorContext(ctx, "failed to unmarshal settlement failed event", "error", err)
			return err
		}
		return h.projector.OnSettlementFailed(ctx, &event)
	default:
		h.logger.WarnContext(ctx, "unknown settlement event topic", "topic", msg.Topic)
		return nil
	}
}
