// 生成摘要：新增支付事件消费处理器，用于驱动读模型投影更新。
// 假设：Kafka topic 与事件类型一一对应，消息体为 JSON。
package consumer

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/segmentio/kafka-go"
	"github.com/wyfcoding/ecommerce/internal/payment/application"
	"github.com/wyfcoding/ecommerce/internal/payment/domain"
)

// PaymentProjectionHandler 处理支付事件并更新读模型。
type PaymentProjectionHandler struct {
	projector *application.PaymentProjectionService
	logger    *slog.Logger
}

// NewPaymentProjectionHandler 创建事件消费处理器。
func NewPaymentProjectionHandler(projector *application.PaymentProjectionService, logger *slog.Logger) *PaymentProjectionHandler {
	return &PaymentProjectionHandler{
		projector: projector,
		logger:    logger,
	}
}

// Handle 处理 Kafka 消息并触发读模型投影。
func (h *PaymentProjectionHandler) Handle(ctx context.Context, msg kafka.Message) error {
	switch msg.Topic {
	case "payment.initiated":
		var event domain.PaymentInitiatedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			h.logger.ErrorContext(ctx, "failed to unmarshal payment initiated event", "error", err)
			return err
		}
		return h.projector.OnPaymentInitiated(ctx, &event)
	case "payment.authorized":
		var event domain.PaymentAuthorizedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			h.logger.ErrorContext(ctx, "failed to unmarshal payment authorized event", "error", err)
			return err
		}
		return h.projector.OnPaymentAuthorized(ctx, &event)
	case "payment.captured":
		var event domain.PaymentCapturedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			h.logger.ErrorContext(ctx, "failed to unmarshal payment captured event", "error", err)
			return err
		}
		return h.projector.OnPaymentCaptured(ctx, &event)
	case "payment.paid":
		var event domain.PaymentPaidEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			h.logger.ErrorContext(ctx, "failed to unmarshal payment paid event", "error", err)
			return err
		}
		return h.projector.OnPaymentPaid(ctx, &event)
	case "payment.refunded":
		var event domain.RefundFinishedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			h.logger.ErrorContext(ctx, "failed to unmarshal refund finished event", "error", err)
			return err
		}
		return h.projector.OnRefundFinished(ctx, &event)
	case "payment.closed":
		var event domain.PaymentClosedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			h.logger.ErrorContext(ctx, "failed to unmarshal payment closed event", "error", err)
			return err
		}
		return h.projector.OnPaymentClosed(ctx, &event)
	default:
		h.logger.WarnContext(ctx, "unknown payment event topic", "topic", msg.Topic)
		return nil
	}
}
