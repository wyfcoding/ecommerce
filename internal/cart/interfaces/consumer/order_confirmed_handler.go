// 生成摘要：新增订单确认事件消费处理器，用于下单后清理购物车。
package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/wyfcoding/ecommerce/internal/cart/application"
	"github.com/wyfcoding/pkg/idempotency"
)

// OrderConfirmedHandler 处理订单确认事件，清理购物车。
type OrderConfirmedHandler struct {
	cmd     *application.CartCommandService
	idem    idempotency.Manager
	logger  *slog.Logger
	idemTTL time.Duration
}

// NewOrderConfirmedHandler 创建订单确认事件处理器。
func NewOrderConfirmedHandler(cmd *application.CartCommandService, idem idempotency.Manager, logger *slog.Logger) *OrderConfirmedHandler {
	return &OrderConfirmedHandler{
		cmd:     cmd,
		idem:    idem,
		logger:  logger,
		idemTTL: 24 * time.Hour,
	}
}

// Handle 处理 Kafka 消息并清理购物车。
func (h *OrderConfirmedHandler) Handle(ctx context.Context, msg kafka.Message) error {
	var event struct {
		OrderID uint64 `json:"order_id"`
		UserID  uint64 `json:"user_id"`
		Items   []struct {
			SkuID uint64 `json:"sku_id"`
		} `json:"items"`
	}
	if err := json.Unmarshal(msg.Value, &event); err != nil {
		h.logger.ErrorContext(ctx, "failed to unmarshal order confirmed event", "error", err)
		return err
	}
	if event.OrderID == 0 || event.UserID == 0 {
		return nil
	}

	idemKey := fmt.Sprintf("cart:clear:order:%d", event.OrderID)
	isFirst, _, err := h.idem.TryStart(ctx, idemKey, h.idemTTL)
	if err != nil || !isFirst {
		return err
	}

	skuIDs := make([]string, 0, len(event.Items))
	for _, it := range event.Items {
		skuIDs = append(skuIDs, fmt.Sprintf("%d", it.SkuID))
	}

	if err := h.cmd.RemoveItems(ctx, event.UserID, skuIDs); err != nil {
		_ = h.idem.Delete(ctx, idemKey)
		return err
	}

	_ = h.idem.Finish(ctx, idemKey, &idempotency.Response{Body: "OK"}, h.idemTTL)
	return nil
}
