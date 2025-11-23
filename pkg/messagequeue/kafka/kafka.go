package kafka

import (
	"context"
	"time"

	"ecommerce/pkg/config"
	"ecommerce/pkg/logging"

	"github.com/segmentio/kafka-go"
)

// Producer wraps kafka.Writer
type Producer struct {
	writer *kafka.Writer
	logger *logging.Logger
}

// NewProducer creates a new Kafka producer.
func NewProducer(cfg config.KafkaConfig, logger *logging.Logger) *Producer {
	w := &kafka.Writer{
		Addr:         kafka.TCP(cfg.Brokers...),
		Topic:        cfg.Topic,
		Balancer:     &kafka.LeastBytes{},
		WriteTimeout: cfg.WriteTimeout,
		ReadTimeout:  cfg.ReadTimeout,
		BatchSize:    cfg.MaxBytes, // Using MaxBytes as batch size approximation or use dedicated config
		BatchTimeout: 10 * time.Millisecond,
		Async:        cfg.Async,
		RequiredAcks: kafka.RequiredAcks(cfg.RequiredAcks),
	}

	return &Producer{
		writer: w,
		logger: logger,
	}
}

// Publish sends a message to Kafka.
func (p *Producer) Publish(ctx context.Context, key, value []byte) error {
	err := p.writer.WriteMessages(ctx, kafka.Message{
		Key:   key,
		Value: value,
		Time:  time.Now(),
	})
	if err != nil {
		p.logger.ErrorContext(ctx, "failed to publish message", "error", err)
		return err
	}
	return nil
}

// Close closes the producer.
func (p *Producer) Close() error {
	return p.writer.Close()
}

// Consumer wraps kafka.Reader
type Consumer struct {
	reader *kafka.Reader
	logger *logging.Logger
}

// NewConsumer creates a new Kafka consumer.
func NewConsumer(cfg config.KafkaConfig, logger *logging.Logger) *Consumer {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:         cfg.Brokers,
		GroupID:         cfg.GroupID,
		Topic:           cfg.Topic,
		MinBytes:        cfg.MinBytes,
		MaxBytes:        cfg.MaxBytes,
		MaxWait:         cfg.MaxWait,
		ReadLagInterval: -1,
		CommitInterval:  cfg.CommitInterval,
		StartOffset:     kafka.LastOffset,
	})

	return &Consumer{
		reader: r,
		logger: logger,
	}
}

// Consume reads messages from Kafka.
func (c *Consumer) Consume(ctx context.Context, handler func(ctx context.Context, msg kafka.Message) error) error {
	for {
		m, err := c.reader.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			c.logger.Error("failed to fetch message", "error", err)
			continue
		}

		// Process message
		if err := handler(ctx, m); err != nil {
			c.logger.Error("handler failed", "error", err)
			// Decide whether to commit or not based on error type
			// For now, we assume manual commit is needed if auto-commit is disabled,
			// but kafka-go Reader with CommitInterval > 0 does auto-commit.
			// If we used FetchMessage, we need to CommitMessages manually.
			continue
		}

		if err := c.reader.CommitMessages(ctx, m); err != nil {
			c.logger.Error("failed to commit message", "error", err)
		}
	}
}

// Close closes the consumer.
func (c *Consumer) Close() error {
	return c.reader.Close()
}
