package consumer

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/segmentio/kafka-go"
	"github.com/wyfcoding/ecommerce/internal/inventory/application"
	"github.com/wyfcoding/ecommerce/internal/inventory/domain"
)

// InventoryProjectionHandler 处理库存事件并更新读模型。
type InventoryProjectionHandler struct {
	projector *application.InventoryProjectionService
	logger    *slog.Logger
}

// NewInventoryProjectionHandler 创建事件消费处理器。
func NewInventoryProjectionHandler(projector *application.InventoryProjectionService, logger *slog.Logger) *InventoryProjectionHandler {
	return &InventoryProjectionHandler{projector: projector, logger: logger}
}

// Handle 处理 Kafka 消息并触发读模型投影。
func (h *InventoryProjectionHandler) Handle(ctx context.Context, msg kafka.Message) error {
	switch msg.Topic {
	case domain.StockLockedEventType:
		var event domain.StockLockedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			h.logger.ErrorContext(ctx, "failed to unmarshal stock locked event", "error", err)
			return err
		}
		return h.projector.OnStockLocked(ctx, &event)
	case domain.StockUnlockedEventType:
		var event domain.StockUnlockedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			h.logger.ErrorContext(ctx, "failed to unmarshal stock unlocked event", "error", err)
			return err
		}
		return h.projector.OnStockUnlocked(ctx, &event)
	case domain.StockDeductedEventType:
		var event domain.StockDeductedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			h.logger.ErrorContext(ctx, "failed to unmarshal stock deducted event", "error", err)
			return err
		}
		return h.projector.OnStockDeducted(ctx, &event)
	case domain.StockAddedEventType:
		var event domain.StockAddedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			h.logger.ErrorContext(ctx, "failed to unmarshal stock added event", "error", err)
			return err
		}
		return h.projector.OnStockAdded(ctx, &event)
	case domain.StockWarningEventType:
		var event domain.StockWarningEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			h.logger.ErrorContext(ctx, "failed to unmarshal stock warning event", "error", err)
			return err
		}
		return h.projector.OnStockWarning(ctx, &event)
	default:
		h.logger.WarnContext(ctx, "unknown inventory event topic", "topic", msg.Topic)
		return nil
	}
}
