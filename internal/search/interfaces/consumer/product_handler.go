// 生成摘要：实现商品事件消费者，监听商品生命周期事件并同步 ES 索引
// 支持针对不同事件类型进行精细化映射，确保搜索数据最终一致性

package consumer

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/segmentio/kafka-go"
	"github.com/wyfcoding/ecommerce/internal/search/application"
)

type ProductEventHandler struct {
	searchService *application.Search
	logger        *slog.Logger
}

func NewProductEventHandler(searchService *application.Search, logger *slog.Logger) *ProductEventHandler {
	return &ProductEventHandler{
		searchService: searchService,
		logger:        logger.With("module", "product_event_handler"),
	}
}

// Handle 处理商品相关的所有 Kafka 消息
func (h *ProductEventHandler) Handle(ctx context.Context, msg kafka.Message) error {
	h.logger.Debug("received product event", "topic", msg.Topic, "partition", msg.Partition, "offset", msg.Offset)

	var payload map[string]any
	if err := json.Unmarshal(msg.Value, &payload); err != nil {
		h.logger.Error("failed to unmarshal product event", "topic", msg.Topic, "error", err)
		return err
	}

	// 将原始事件转换为 SearchManager 可以理解的同步格式
	syncEvent := map[string]any{
		"product_id": payload["id"],
		"timestamp":  payload["timestamp"],
	}

	switch msg.Topic {
	case "product.created":
		syncEvent["action"] = "create"
		syncEvent["name"] = payload["name"]
		syncEvent["price"] = payload["price"]
		syncEvent["stock"] = payload["stock"]
		syncEvent["category_id"] = payload["category_id"]
		syncEvent["category_name"] = payload["category_name"]
		syncEvent["brand_id"] = payload["brand_id"]
		syncEvent["brand_name"] = payload["brand_name"]
		syncEvent["image_url"] = payload["image_url"]
		syncEvent["description"] = payload["description"]
	case "product.updated":
		syncEvent["action"] = "update"
		// 动态复制所有 payload 到 syncEvent
		for k, v := range payload {
			if k != "id" && k != "timestamp" {
				syncEvent[k] = v
			}
		}
	case "product.inventory_changed":
		syncEvent["action"] = "update"
		syncEvent["stock"] = payload["available_stock"]
	case "product.deleted":
		syncEvent["action"] = "delete"
	default:
		h.logger.Warn("unsupported product topic", "topic", msg.Topic)
		return nil
	}

	if err := h.searchService.SyncProductIndex(ctx, syncEvent); err != nil {
		h.logger.Error("failed to sync product index", "product_id", syncEvent["product_id"], "error", err)
		return err
	}

	return nil
}
