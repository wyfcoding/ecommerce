package events

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/wyfcoding/ecommerce/internal/product/domain"
)

// ProductSearchHandler 监听商品变更事件并同步到 Elasticsearch。
type ProductSearchHandler struct {
	searchRepo domain.ProductSearchRepository
}

func NewProductSearchHandler(searchRepo domain.ProductSearchRepository) *ProductSearchHandler {
	return &ProductSearchHandler{searchRepo: searchRepo}
}

// HandleProductCreated 处理商品创建事件。
func (h *ProductSearchHandler) HandleProductCreated(ctx context.Context, payload []byte) error {
	var event domain.ProductCreatedEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		return err
	}

	product := &domain.Product{
		Name:  event.Name,
		Price: event.Price,
		Stock: event.Stock,
	}
	product.ID = event.ID

	return h.searchRepo.Index(ctx, product)
}

// HandleProductUpdated 处理商品更新事件。
func (h *ProductSearchHandler) HandleProductUpdated(ctx context.Context, payload []byte) error {
	var event domain.ProductUpdatedEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		return err
	}

	product := &domain.Product{
		Status: domain.ProductStatus(event.Status),
	}
	product.ID = event.ID

	return h.searchRepo.Index(ctx, product)
}

// HandleProductDeleted 处理商品删除事件。
func (h *ProductSearchHandler) HandleProductDeleted(ctx context.Context, payload []byte) error {
	var event domain.ProductDeletedEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		return err
	}

	return h.searchRepo.Delete(ctx, uint64(event.ID))
}

// Subscribe 将处理程序注册到消费者。
func (h *ProductSearchHandler) Subscribe(ctx context.Context, consumer any) {
	// 实际代码中这里会调用具体消息队列的 Subscribe 方法
	slog.Info("Consuming product events for ES indexing")
}
