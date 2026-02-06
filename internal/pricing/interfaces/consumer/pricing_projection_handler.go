// 生成摘要：新增定价事件消费处理器，用于驱动读模型投影更新。
package consumer

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/segmentio/kafka-go"
	"github.com/wyfcoding/ecommerce/internal/pricing/application"
	"github.com/wyfcoding/ecommerce/internal/pricing/domain"
)

// PricingProjectionHandler 处理定价事件并更新读模型。
type PricingProjectionHandler struct {
	projector *application.PricingProjectionService
	logger    *slog.Logger
}

// NewPricingProjectionHandler 创建事件消费处理器。
func NewPricingProjectionHandler(projector *application.PricingProjectionService, logger *slog.Logger) *PricingProjectionHandler {
	return &PricingProjectionHandler{
		projector: projector,
		logger:    logger,
	}
}

// Handle 处理 Kafka 消息并触发读模型投影。
func (h *PricingProjectionHandler) Handle(ctx context.Context, msg kafka.Message) error {
	switch msg.Topic {
	case domain.PricingRuleUpdatedEventType:
		var event domain.PricingRuleUpdatedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			h.logger.ErrorContext(ctx, "failed to unmarshal pricing rule updated event", "error", err)
			return err
		}
		return h.projector.OnPricingRuleUpdated(ctx, &event)
	case domain.PriceHistoryRecordedEventType:
		var event domain.PriceHistoryRecordedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			h.logger.ErrorContext(ctx, "failed to unmarshal price history recorded event", "error", err)
			return err
		}
		return h.projector.OnPriceHistoryRecorded(ctx, &event)
	default:
		h.logger.WarnContext(ctx, "unknown pricing event topic", "topic", msg.Topic)
		return nil
	}
}
