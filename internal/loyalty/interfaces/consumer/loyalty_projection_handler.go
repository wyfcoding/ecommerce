// 生成摘要：新增忠诚度事件消费处理器，用于驱动读模型投影更新。
// 假设：Kafka topic 与事件类型一一对应，消息体为 JSON。
package consumer

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/segmentio/kafka-go"
	"github.com/wyfcoding/ecommerce/internal/loyalty/application"
	"github.com/wyfcoding/ecommerce/internal/loyalty/domain"
)

// LoyaltyProjectionHandler 处理忠诚度事件并更新读模型。
type LoyaltyProjectionHandler struct {
	projector *application.LoyaltyProjectionService
	logger    *slog.Logger
}

// NewLoyaltyProjectionHandler 创建事件消费处理器。
func NewLoyaltyProjectionHandler(projector *application.LoyaltyProjectionService, logger *slog.Logger) *LoyaltyProjectionHandler {
	return &LoyaltyProjectionHandler{
		projector: projector,
		logger:    logger,
	}
}

// Handle 处理 Kafka 消息并触发读模型投影。
func (h *LoyaltyProjectionHandler) Handle(ctx context.Context, msg kafka.Message) error {
	switch msg.Topic {
	case domain.MemberAccountUpdatedEventType:
		var event domain.MemberAccountUpdatedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			h.logger.ErrorContext(ctx, "failed to unmarshal account updated event", "error", err)
			return err
		}
		return h.projector.OnAccountUpdated(ctx, &event)
	case domain.PointsTransactionCreatedEventType:
		var event domain.PointsTransactionCreatedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			h.logger.ErrorContext(ctx, "failed to unmarshal transaction created event", "error", err)
			return err
		}
		return h.projector.OnTransactionCreated(ctx, &event)
	case domain.MemberBenefitSavedEventType:
		var event domain.MemberBenefitSavedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			h.logger.ErrorContext(ctx, "failed to unmarshal benefit saved event", "error", err)
			return err
		}
		return h.projector.OnBenefitSaved(ctx, &event)
	case domain.MemberBenefitDeletedEventType:
		var event domain.MemberBenefitDeletedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			h.logger.ErrorContext(ctx, "failed to unmarshal benefit deleted event", "error", err)
			return err
		}
		return h.projector.OnBenefitDeleted(ctx, &event)
	default:
		h.logger.WarnContext(ctx, "unknown loyalty event topic", "topic", msg.Topic)
		return nil
	}
}
