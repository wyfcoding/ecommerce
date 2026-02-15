// 生成摘要：
// - 实现基于 Kafka 的事件发布者，对接 pkg/messagequeue 接口
// - 将领域事件转换为 Kafka 消息并推送至指定主题
// - 集成日志记录，以便追踪事件发布状态

package mq

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/wyfcoding/pkg/messagequeue"
	"github.com/wyfcoding/pkg/messagequeue/kafka"
)

// KafkaEventPublisher 基于 Kafka 的事件发布者实现
type KafkaEventPublisher struct {
	producer *kafka.Producer
	logger   *slog.Logger
}

// NewKafkaEventPublisher 创建事件发布者
func NewKafkaEventPublisher(producer *kafka.Producer, logger *slog.Logger) messagequeue.EventPublisher {
	return &KafkaEventPublisher{
		producer: producer,
		logger:   logger.With("module", "kafka_event_publisher"),
	}
}

// Publish 发布事件
func (p *KafkaEventPublisher) Publish(ctx context.Context, topic, key string, event any) error {
	if p.producer == nil {
		p.logger.WarnContext(ctx, "kafka producer is nil, skipping publish", "topic", topic)
		return nil
	}

	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}

	if err := p.producer.PublishToTopic(ctx, topic, []byte(key), payload); err != nil {
		p.logger.ErrorContext(ctx, "failed to publish event to kafka",
			"topic", topic,
			"error", err)
		return err
	}

	return nil
}

// PublishInTx 在事务中发布事件（当前简单包装，未来可扩展 Outbox）
func (p *KafkaEventPublisher) PublishInTx(ctx context.Context, tx any, topic, key string, event any) error {
	// TODO: 集成 Outbox 模式以支持真正的事务一致性
	return p.Publish(ctx, topic, key, event)
}
