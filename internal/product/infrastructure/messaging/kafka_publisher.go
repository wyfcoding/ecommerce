package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/segmentio/kafka-go"
)

type KafkaPublisher struct {
	writer *kafka.Writer
	logger *slog.Logger
}

// NewKafkaPublisher 创建 Kafka 消息发布者
func NewKafkaPublisher(brokers []string, logger *slog.Logger) *KafkaPublisher {
	w := &kafka.Writer{
		Addr:         kafka.TCP(brokers...),
		Balancer:     &kafka.LeastBytes{},
		BatchSize:    100,
		BatchTimeout: 10 * time.Millisecond,
		Async:        true, // 异步发送提高性能
		Completion: func(messages []kafka.Message, err error) {
			if err != nil {
				logger.Error("kafka write failed", "error", err)
			}
		},
	}

	return &KafkaPublisher{
		writer: w,
		logger: logger,
	}
}

// Publish 发布事件到指定 Topic
func (p *KafkaPublisher) Publish(ctx context.Context, topic string, event interface{}) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	// 尝试从事件中提取 Key (例如 UserID) 以保证有序性，这里简化为随机(无Key)或由Balancer决定
	// 如果需要严格顺序，需要定义 Key。这里假设 event 结构体没有通用 Key 接口。
	// 改进：若 event 实现了 Key() string 接口则使用。

	msg := kafka.Message{
		Topic: topic,
		Value: payload,
		Time:  time.Now(),
	}

	// 使用 writer 写入。注意：如果是 Async=true，WriteMessages 可能只进入 buffer。
	if err := p.writer.WriteMessages(ctx, msg); err != nil {
		return fmt.Errorf("failed to publish message to kafka: %w", err)
	}

	p.logger.DebugContext(ctx, "event published", "topic", topic, "payload_size", len(payload))
	return nil
}

// Close 关闭连接
func (p *KafkaPublisher) Close() error {
	return p.writer.Close()
}
