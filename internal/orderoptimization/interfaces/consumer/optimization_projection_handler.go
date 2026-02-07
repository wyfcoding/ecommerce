// 生成摘要：新增订单优化事件消费处理器，用于驱动读模型投影更新。
// 假设：Kafka topic 与事件类型一一对应，消息体为 JSON。
package consumer

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/segmentio/kafka-go"
	"github.com/wyfcoding/ecommerce/internal/orderoptimization/application"
	"github.com/wyfcoding/ecommerce/internal/orderoptimization/domain"
)

// OptimizationProjectionHandler 处理订单优化事件并更新读模型。
type OptimizationProjectionHandler struct {
	projector *application.OptimizationProjectionService
	logger    *slog.Logger
}

// NewOptimizationProjectionHandler 创建事件消费处理器。
func NewOptimizationProjectionHandler(projector *application.OptimizationProjectionService, logger *slog.Logger) *OptimizationProjectionHandler {
	return &OptimizationProjectionHandler{
		projector: projector,
		logger:    logger,
	}
}

// Handle 处理 Kafka 消息并触发读模型投影。
func (h *OptimizationProjectionHandler) Handle(ctx context.Context, msg kafka.Message) error {
	switch msg.Topic {
	case domain.OrderMergedEventType:
		var event domain.OrderMergedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			h.logger.ErrorContext(ctx, "failed to unmarshal order merged event", "error", err)
			return err
		}
		return h.projector.OnOrderMerged(ctx, &event)
	case domain.OrderSplitEventType:
		var event domain.OrderSplitEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			h.logger.ErrorContext(ctx, "failed to unmarshal order split event", "error", err)
			return err
		}
		return h.projector.OnOrderSplit(ctx, &event)
	case domain.OrderAllocationPlanCreatedType:
		var event domain.OrderAllocationPlanCreatedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			h.logger.ErrorContext(ctx, "failed to unmarshal allocation plan created event", "error", err)
			return err
		}
		return h.projector.OnAllocationPlanCreated(ctx, &event)
	default:
		h.logger.WarnContext(ctx, "unknown order optimization event topic", "topic", msg.Topic)
		return nil
	}
}
