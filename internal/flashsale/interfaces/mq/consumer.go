package mq

import (
	"context"
	"encoding/json"
	"log/slog"

	kafkago "github.com/segmentio/kafka-go"
	"github.com/wyfcoding/ecommerce/internal/flashsale/domain"
	"github.com/wyfcoding/pkg/messagequeue/kafka"
)

// OrderConsumer 负责异步处理秒杀订单创建。
type OrderConsumer struct {
	consumer *kafka.Consumer
	repo     domain.FlashSaleRepository
	logger   *slog.Logger
}

// NewOrderConsumer 创建 OrderConsumer 实例。
func NewOrderConsumer(consumer *kafka.Consumer, repo domain.FlashSaleRepository, logger *slog.Logger) *OrderConsumer {
	return &OrderConsumer{
		consumer: consumer,
		repo:     repo,
		logger:   logger,
	}
}

// Start 启动消费者协程。
func (c *OrderConsumer) Start(ctx context.Context) error {
	c.logger.Info("Starting OrderConsumer with concurrent workers...")
	// 使用优化后的 Start 方法，开启 5 个并发 Worker 进行消费。
	c.consumer.Start(ctx, 5, c.handleMessage)
	return nil
}

// handleMessage 处理单条 Kafka 消息。
func (c *OrderConsumer) handleMessage(ctx context.Context, msg kafkago.Message) error {
	var event map[string]any
	if err := json.Unmarshal(msg.Value, &event); err != nil {
		c.logger.Error("failed to unmarshal message", "error", err)
		return nil // 格式错误的消息直接跳过，不重试
	}

	// 提取事件字段 (增加防御性检查)
	val, ok := event["order_id"]
	if !ok {
		return nil
	}
	orderID := uint64(val.(float64))

	flashsaleID := uint64(event["flashsale_id"].(float64))
	userID := uint64(event["user_id"].(float64))
	productID := uint64(event["product_id"].(float64))
	skuID := uint64(event["sku_id"].(float64))
	quantity := int32(event["quantity"].(float64))
	price := int64(event["price"].(float64))

	order := domain.NewFlashsaleOrder(flashsaleID, userID, productID, skuID, quantity, price)
	order.ID = uint(orderID)
	order.Status = domain.FlashsaleOrderStatusPending

	// 幂等检查：防止重复保存
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

// Stop 停止消费者并释放资源。
func (c *OrderConsumer) Stop(ctx context.Context) error {
	c.logger.Info("Stopping OrderConsumer...")
	return c.consumer.Close()
}
