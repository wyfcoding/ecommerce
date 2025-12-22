package mq

import (
	"context"
	"encoding/json"

	"log/slog"

	kafkago "github.com/segmentio/kafka-go"
	"github.com/wyfcoding/ecommerce/internal/flashsale/domain"
	"github.com/wyfcoding/pkg/messagequeue/kafka"
)

// OrderConsumer 结构体定义。
type OrderConsumer struct {
	consumer *kafka.Consumer
	repo     domain.FlashSaleRepository
	logger   *slog.Logger
}

// NewOrderConsumer 函数。
func NewOrderConsumer(consumer *kafka.Consumer, repo domain.FlashSaleRepository, logger *slog.Logger) *OrderConsumer {
	return &OrderConsumer{
		consumer: consumer,
		repo:     repo,
		logger:   logger,
	}
}

func (c *OrderConsumer) Start(ctx context.Context) error {
	c.logger.Info("Starting OrderConsumer...")
	return c.consumer.Consume(ctx, c.handleMessage)
}

func (c *OrderConsumer) handleMessage(ctx context.Context, msg kafkago.Message) error {
	var event map[string]any
	if err := json.Unmarshal(msg.Value, &event); err != nil {
		c.logger.Error("failed to unmarshal message", "error", err)
		return nil
	}

	orderID := uint64(event["order_id"].(float64))
	flashsaleID := uint64(event["flashsale_id"].(float64))
	userID := uint64(event["user_id"].(float64))
	productID := uint64(event["product_id"].(float64))
	skuID := uint64(event["sku_id"].(float64))
	quantity := int32(event["quantity"].(float64))
	price := int64(event["price"].(float64))

	order := domain.NewFlashsaleOrder(flashsaleID, userID, productID, skuID, quantity, price)
	order.ID = uint(orderID)
	order.Status = domain.FlashsaleOrderStatusPending

	existing, err := c.repo.GetOrder(ctx, orderID)
	if err == nil && existing != nil {
		c.logger.Info("order already exists, skipping", "order_id", orderID)
		return nil
	}

	if err := c.repo.SaveOrder(ctx, order); err != nil {
		c.logger.Error("failed to save order from mq", "error", err)
		return err
	}

	c.logger.Info("successfully created order from mq", "order_id", orderID)
	return nil
}

func (c *OrderConsumer) Stop(ctx context.Context) error {
	c.logger.Info("Stopping OrderConsumer...")
	return c.consumer.Close()
}
