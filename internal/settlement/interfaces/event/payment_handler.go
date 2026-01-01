package event

import (
	"context"
	"encoding/json"
	"log/slog"

	kafkago "github.com/segmentio/kafka-go"
	"github.com/wyfcoding/ecommerce/internal/settlement/application"
)

// PaymentHandler 处理支付相关的消息事件。
type PaymentHandler struct {
	app    *application.SettlementService
	logger *slog.Logger
}

// NewPaymentHandler 创建支付事件处理器。
func NewPaymentHandler(app *application.SettlementService, logger *slog.Logger) *PaymentHandler {
	return &PaymentHandler{
		app:    app,
		logger: logger.With("module", "event_handler"),
	}
}

// HandlePaymentCaptured 处理支付捕获事件，将其记录到结算单和账本中。
func (h *PaymentHandler) HandlePaymentCaptured(ctx context.Context, msg kafkago.Message) error {
	var event struct {
		PaymentNo string `json:"payment_no"`
		OrderNo   string `json:"order_no"`
		UserID    uint64 `json:"user_id"`
		Amount    int64  `json:"amount"`
		Timestamp int64  `json:"timestamp"`
	}

	if err := json.Unmarshal(msg.Value, &event); err != nil {
		h.logger.Error("failed to unmarshal payment captured event", "error", err)
		return err
	}

	h.logger.Info("received payment captured event", "payment_no", event.PaymentNo, "amount", event.Amount)

	// 业务逻辑：记录支付成功，更新结算单
	// 假设 merchant_id 为 1 (实际应从订单或产品信息中获取)
	merchantID := uint64(1)
	channelCost := int64(float64(event.Amount) * 0.006) // 假设 0.6% 渠道成本

	return h.app.RecordPaymentSuccess(ctx, 0, event.OrderNo, merchantID, event.Amount, channelCost)
}
