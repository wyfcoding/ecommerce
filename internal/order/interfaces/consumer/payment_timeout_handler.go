package consumer

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/wyfcoding/ecommerce/internal/order/application"
	"github.com/wyfcoding/ecommerce/internal/order/domain"
)

// PaymentTimeoutHandler 处理订单支付超时事件。
type PaymentTimeoutHandler struct {
	orderCmd  *application.OrderCommandService
	scheduler domain.TimeoutScheduler
	logger    *slog.Logger
}

// NewPaymentTimeoutHandler 构造函数。
func NewPaymentTimeoutHandler(orderCmd *application.OrderCommandService, scheduler domain.TimeoutScheduler, logger *slog.Logger) *PaymentTimeoutHandler {
	return &PaymentTimeoutHandler{
		orderCmd:  orderCmd,
		scheduler: scheduler,
		logger:    logger,
	}
}

// Handle 消费支付超时事件并调度取消。
func (h *PaymentTimeoutHandler) Handle(ctx context.Context, msg kafka.Message) error {
	var event domain.OrderPaymentTimeoutEvent
	if err := json.Unmarshal(msg.Value, &event); err != nil {
		h.logger.ErrorContext(ctx, "failed to unmarshal payment timeout event", "error", err)
		return err
	}

	expiresAt := time.Unix(event.ExpiresAt, 0)
	delay := time.Until(expiresAt)
	if delay <= 0 || h.scheduler == nil {
		h.cancelIfPending(event)
		return nil
	}

	return h.scheduler.ScheduleTimeout(event.OrderNo, delay, func(_ string) {
		h.cancelIfPending(event)
	})
}

func (h *PaymentTimeoutHandler) cancelIfPending(event domain.OrderPaymentTimeoutEvent) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := h.orderCmd.CancelOrderIfPending(ctx, &application.CancelOrderCommand{
		UserID:   event.UserID,
		OrderID:  event.OrderID,
		Operator: "System",
		Reason:   "payment timeout",
	}); err != nil {
		h.logger.WarnContext(ctx, "failed to cancel order on payment timeout", "order_id", event.OrderID, "error", err)
	}
}
