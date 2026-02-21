// Package consumer 处理库存事件
package consumer

import (
	"context"
	"encoding/json"

	"github.com/segmentio/kafka-go"
	"github.com/wyfcoding/ecommerce/internal/order/application"
)

type InventoryEventConsumer struct {
	orderApp *application.OrderCommandService
}

func (c *InventoryEventConsumer) Handle(ctx context.Context, msg kafka.Message) error {
	var event struct {
		OrderID string `json:"order_id"`
		Status  string `json:"status"` // SUCCESS, OUT_OF_STOCK
	}
	json.Unmarshal(msg.Value, &event)

	if event.Status == "SUCCESS" {
		return c.orderApp.MarkInventoryAllocated(ctx, event.OrderID)
	} else {
		return c.orderApp.MarkOutOfStock(ctx, event.OrderID)
	}
}
