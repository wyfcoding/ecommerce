package consumer

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/segmentio/kafka-go"
	"github.com/wyfcoding/ecommerce/internal/dynamicpricing/application"
	"github.com/wyfcoding/ecommerce/internal/dynamicpricing/domain"
)

// DynamicPricingProjectionHandler 处理定价事件并更新读模型。
type DynamicPricingProjectionHandler struct {
	projector *application.DynamicPricingProjectionService
	logger    *slog.Logger
}

// NewDynamicPricingProjectionHandler 创建事件消费处理器。
func NewDynamicPricingProjectionHandler(projector *application.DynamicPricingProjectionService, logger *slog.Logger) *DynamicPricingProjectionHandler {
	return &DynamicPricingProjectionHandler{
		projector: projector,
		logger:    logger,
	}
}

// Handle 处理 Kafka 消息并触发读模型投影。
func (h *DynamicPricingProjectionHandler) Handle(ctx context.Context, msg kafka.Message) error {
	switch msg.Topic {
	case domain.DynamicPriceCalculatedEventType:
		var event domain.PriceCalculatedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			h.logger.ErrorContext(ctx, "failed to unmarshal price calculated event", "error", err)
			return err
		}
		return h.projector.OnPriceCalculated(ctx, &event)
	case domain.PricingStrategyCreatedEventType, domain.PricingStrategyUpdatedEventType:
		var event domain.StrategyUpdatedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			h.logger.ErrorContext(ctx, "failed to unmarshal strategy updated event", "error", err)
			return err
		}
		return h.projector.OnStrategyChanged(ctx, &event)
	default:
		h.logger.WarnContext(ctx, "unknown dynamicpricing event topic", "topic", msg.Topic)
		return nil
	}
}
