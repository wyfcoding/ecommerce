// 生成摘要：新增积分商城事件消费处理器，用于驱动读模型投影更新。
// 假设：Kafka topic 与事件类型一一对应，消息体为 JSON。
package consumer

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/segmentio/kafka-go"
	"github.com/wyfcoding/ecommerce/internal/pointsmall/application"
	"github.com/wyfcoding/ecommerce/internal/pointsmall/domain"
)

// PointsmallProjectionHandler 处理积分商城事件并更新读模型。
type PointsmallProjectionHandler struct {
	projector *application.PointsmallProjectionService
	logger    *slog.Logger
}

// NewPointsmallProjectionHandler 创建事件消费处理器。
func NewPointsmallProjectionHandler(projector *application.PointsmallProjectionService, logger *slog.Logger) *PointsmallProjectionHandler {
	return &PointsmallProjectionHandler{
		projector: projector,
		logger:    logger,
	}
}

// Handle 处理 Kafka 消息并触发读模型投影。
func (h *PointsmallProjectionHandler) Handle(ctx context.Context, msg kafka.Message) error {
	switch msg.Topic {
	case domain.PointsProductCreatedEventType:
		var event domain.PointsProductCreatedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			h.logger.ErrorContext(ctx, "failed to unmarshal points product created event", "error", err)
			return err
		}
		return h.projector.OnProductCreated(ctx, &event)
	case domain.PointsStockUpdatedEventType:
		var event domain.PointsStockUpdatedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			h.logger.ErrorContext(ctx, "failed to unmarshal points stock updated event", "error", err)
			return err
		}
		return h.projector.OnStockUpdated(ctx, &event)
	case domain.PointsOrderCreatedEventType:
		var event domain.PointsOrderCreatedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			h.logger.ErrorContext(ctx, "failed to unmarshal points order created event", "error", err)
			return err
		}
		return h.projector.OnOrderCreated(ctx, &event)
	case domain.PointsAccountUpdatedEventType:
		var event domain.PointsAccountUpdatedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			h.logger.ErrorContext(ctx, "failed to unmarshal points account updated event", "error", err)
			return err
		}
		return h.projector.OnAccountUpdated(ctx, &event)
	default:
		h.logger.WarnContext(ctx, "unknown pointsmall event topic", "topic", msg.Topic)
		return nil
	}
}
