package consumer

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/segmentio/kafka-go"
	"github.com/wyfcoding/ecommerce/internal/warehouse/application"
	"github.com/wyfcoding/ecommerce/internal/warehouse/domain"
)

// WarehouseProjectionHandler 处理仓库事件并更新读模型。
type WarehouseProjectionHandler struct {
	projector *application.WarehouseProjectionService
	logger    *slog.Logger
}

// NewWarehouseProjectionHandler 创建事件消费处理器。
func NewWarehouseProjectionHandler(projector *application.WarehouseProjectionService, logger *slog.Logger) *WarehouseProjectionHandler {
	return &WarehouseProjectionHandler{projector: projector, logger: logger}
}

// Handle 处理 Kafka 消息并触发读模型投影。
func (h *WarehouseProjectionHandler) Handle(ctx context.Context, msg kafka.Message) error {
	switch msg.Topic {
	case domain.WarehouseCreatedEventType:
		var event domain.WarehouseCreatedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			h.logger.ErrorContext(ctx, "failed to unmarshal warehouse created event", "error", err)
			return err
		}
		return h.projector.OnWarehouseCreated(ctx, &event)
	case domain.StockAdjustedEventType:
		var event domain.StockAdjustedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			h.logger.ErrorContext(ctx, "failed to unmarshal stock adjusted event", "error", err)
			return err
		}
		return h.projector.OnStockAdjusted(ctx, &event)
	case domain.StockDeductedEventType:
		var event domain.StockDeductedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			h.logger.ErrorContext(ctx, "failed to unmarshal stock deducted event", "error", err)
			return err
		}
		return h.projector.OnStockDeducted(ctx, &event)
	case domain.StockRevertedEventType:
		var event domain.StockRevertedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			h.logger.ErrorContext(ctx, "failed to unmarshal stock reverted event", "error", err)
			return err
		}
		return h.projector.OnStockReverted(ctx, &event)
	case domain.StockTransferCreatedEventType:
		var event domain.StockTransferCreatedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			h.logger.ErrorContext(ctx, "failed to unmarshal transfer created event", "error", err)
			return err
		}
		return h.projector.OnTransferCreated(ctx, &event)
	case domain.StockTransferCompletedEventType:
		var event domain.StockTransferCompletedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			h.logger.ErrorContext(ctx, "failed to unmarshal transfer completed event", "error", err)
			return err
		}
		return h.projector.OnTransferCompleted(ctx, &event)
	default:
		h.logger.WarnContext(ctx, "unknown warehouse event topic", "topic", msg.Topic)
		return nil
	}
}
