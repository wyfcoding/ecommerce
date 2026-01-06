package event

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/segmentio/kafka-go"
	"github.com/wyfcoding/ecommerce/internal/order/application"
)

// FlashsaleHandler 处理来自秒杀服务的异步订单事件。
type FlashsaleHandler struct {
	orderApp *application.OrderManager
	logger   *slog.Logger
}

// NewFlashsaleHandler 构造函数。
func NewFlashsaleHandler(orderApp *application.OrderManager, logger *slog.Logger) *FlashsaleHandler {
	return &FlashsaleHandler{
		orderApp: orderApp,
		logger:   logger,
	}
}

// FlashsaleOrderEvent 秒杀订单事件载荷。
type FlashsaleOrderEvent struct {
	OrderID     uint64 `json:"order_id"`
	FlashsaleID uint64 `json:"flashsale_id"`
	UserID      uint64 `json:"user_id"`
	ProductID   uint64 `json:"product_id"`
	SkuID       uint64 `json:"sku_id"`
	Quantity    int32  `json:"quantity"`
	Price       int64  `json:"price"`
	CreatedAt   string `json:"created_at"`
}

// HandleFlashsaleOrder 消费来自 Kafka 的秒杀成功事件，执行订单的异步落库操作。
// 核心逻辑：解析 JSON -> 提取关键字段 -> 调用领域服务处理持久化。
func (h *FlashsaleHandler) HandleFlashsaleOrder(ctx context.Context, msg kafka.Message) error {
	h.logger.InfoContext(ctx, "received flashsale order event", "msg_key", string(msg.Key))

	var event FlashsaleOrderEvent
	if err := json.Unmarshal(msg.Value, &event); err != nil {
		h.logger.ErrorContext(ctx, "failed to unmarshal flashsale order event", "error", err)
		return err
	}

	// 触发订单管理服务的落库逻辑（幂等性由底层 Repository 保证）
	if err := h.orderApp.HandleFlashsaleOrder(ctx, event.OrderID, event.UserID, event.ProductID, event.SkuID, event.Quantity, event.Price); err != nil {
		h.logger.ErrorContext(ctx, "failed to process flashsale order persistence", "order_id", event.OrderID, "error", err)
		return err
	}

	h.logger.InfoContext(ctx, "flashsale order event processed successfully", "order_id", event.OrderID)
	return nil
}
