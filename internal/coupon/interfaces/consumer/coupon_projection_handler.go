// 生成摘要：新增优惠券事件消费处理器，用于驱动读模型投影更新。
// 假设：Kafka topic 与事件类型一一对应，消息体为 JSON。
package consumer

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/segmentio/kafka-go"
	"github.com/wyfcoding/ecommerce/internal/coupon/application"
	"github.com/wyfcoding/ecommerce/internal/coupon/domain"
)

// CouponProjectionHandler 处理优惠券事件并更新读模型。
type CouponProjectionHandler struct {
	projector *application.CouponProjectionService
	logger    *slog.Logger
}

// NewCouponProjectionHandler 创建事件消费处理器。
func NewCouponProjectionHandler(projector *application.CouponProjectionService, logger *slog.Logger) *CouponProjectionHandler {
	return &CouponProjectionHandler{
		projector: projector,
		logger:    logger,
	}
}

// Handle 处理 Kafka 消息并触发读模型投影。
func (h *CouponProjectionHandler) Handle(ctx context.Context, msg kafka.Message) error {
	switch msg.Topic {
	case domain.CouponCreatedEventType:
		var event domain.CouponCreatedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			h.logger.ErrorContext(ctx, "failed to unmarshal coupon created event", "error", err)
			return err
		}
		return h.projector.OnCouponCreated(ctx, &event)
	case domain.CouponIssuedEventType:
		var event domain.CouponIssuedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			h.logger.ErrorContext(ctx, "failed to unmarshal coupon issued event", "error", err)
			return err
		}
		return h.projector.OnCouponIssued(ctx, &event)
	case domain.CouponUsedEventType:
		var event domain.CouponUsedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			h.logger.ErrorContext(ctx, "failed to unmarshal coupon used event", "error", err)
			return err
		}
		return h.projector.OnCouponUsed(ctx, &event)
	case domain.CouponExpiredEventType:
		var event domain.CouponExpiredEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			h.logger.ErrorContext(ctx, "failed to unmarshal coupon expired event", "error", err)
			return err
		}
		return h.projector.OnCouponExpired(ctx, &event)
	default:
		h.logger.WarnContext(ctx, "unknown coupon event topic", "topic", msg.Topic)
		return nil
	}
}
