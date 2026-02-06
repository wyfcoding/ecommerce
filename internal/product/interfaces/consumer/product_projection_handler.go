package consumer

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/segmentio/kafka-go"
	"github.com/wyfcoding/ecommerce/internal/product/application"
	"github.com/wyfcoding/ecommerce/internal/product/domain"
)

// ProductProjectionHandler 处理商品事件并更新读模型。
type ProductProjectionHandler struct {
	projector *application.ProductProjectionService
	logger    *slog.Logger
}

// NewProductProjectionHandler 创建事件消费处理器。
func NewProductProjectionHandler(projector *application.ProductProjectionService, logger *slog.Logger) *ProductProjectionHandler {
	return &ProductProjectionHandler{projector: projector, logger: logger}
}

// Handle 处理 Kafka 消息并触发读模型投影。
func (h *ProductProjectionHandler) Handle(ctx context.Context, msg kafka.Message) error {
	switch msg.Topic {
	case domain.ProductCreatedEventType:
		var event domain.ProductCreatedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			h.logger.ErrorContext(ctx, "failed to unmarshal product created event", "error", err)
			return err
		}
		return h.projector.OnProductCreated(ctx, &event)
	case domain.ProductUpdatedEventType:
		var event domain.ProductUpdatedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			h.logger.ErrorContext(ctx, "failed to unmarshal product updated event", "error", err)
			return err
		}
		return h.projector.OnProductUpdated(ctx, &event)
	case domain.ProductDeletedEventType:
		var event domain.ProductDeletedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			h.logger.ErrorContext(ctx, "failed to unmarshal product deleted event", "error", err)
			return err
		}
		return h.projector.OnProductDeleted(ctx, &event)
	case domain.SKUAddedEventType:
		var event domain.SKUAddedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			h.logger.ErrorContext(ctx, "failed to unmarshal sku added event", "error", err)
			return err
		}
		return h.projector.OnSKUAdded(ctx, &event)
	case domain.SKUUpdatedEventType:
		var event domain.SKUUpdatedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			h.logger.ErrorContext(ctx, "failed to unmarshal sku updated event", "error", err)
			return err
		}
		return h.projector.OnSKUUpdated(ctx, &event)
	case domain.SKUDeletedEventType:
		var event domain.SKUDeletedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			h.logger.ErrorContext(ctx, "failed to unmarshal sku deleted event", "error", err)
			return err
		}
		return h.projector.OnSKUDeleted(ctx, &event)
	case domain.BrandCreatedEventType:
		var event domain.BrandCreatedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			h.logger.ErrorContext(ctx, "failed to unmarshal brand created event", "error", err)
			return err
		}
		return h.projector.OnBrandCreated(ctx, &event)
	case domain.BrandUpdatedEventType:
		var event domain.BrandUpdatedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			h.logger.ErrorContext(ctx, "failed to unmarshal brand updated event", "error", err)
			return err
		}
		return h.projector.OnBrandUpdated(ctx, &event)
	case domain.BrandDeletedEventType:
		var event domain.BrandDeletedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			h.logger.ErrorContext(ctx, "failed to unmarshal brand deleted event", "error", err)
			return err
		}
		return h.projector.OnBrandDeleted(ctx, &event)
	case domain.CategoryCreatedEventType:
		var event domain.CategoryCreatedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			h.logger.ErrorContext(ctx, "failed to unmarshal category created event", "error", err)
			return err
		}
		return h.projector.OnCategoryCreated(ctx, &event)
	case domain.CategoryUpdatedEventType:
		var event domain.CategoryUpdatedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			h.logger.ErrorContext(ctx, "failed to unmarshal category updated event", "error", err)
			return err
		}
		return h.projector.OnCategoryUpdated(ctx, &event)
	case domain.CategoryDeletedEventType:
		var event domain.CategoryDeletedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			h.logger.ErrorContext(ctx, "failed to unmarshal category deleted event", "error", err)
			return err
		}
		return h.projector.OnCategoryDeleted(ctx, &event)
	default:
		h.logger.WarnContext(ctx, "unknown product event topic", "topic", msg.Topic)
		return nil
	}
}
