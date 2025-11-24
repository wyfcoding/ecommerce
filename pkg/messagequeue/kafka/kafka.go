package kafka

import (
	"context"
	"fmt"
	"time"

	"ecommerce/pkg/config"
	"ecommerce/pkg/logging"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/segmentio/kafka-go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

var (
	mqProduced = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "mq_produced_total",
			Help: "The total number of messages produced",
		},
		[]string{"topic", "status"},
	)
	mqConsumed = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "mq_consumed_total",
			Help: "The total number of messages consumed",
		},
		[]string{"topic", "status"},
	)
	mqDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "mq_operation_duration_seconds",
			Help:    "The duration of mq operations",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"topic", "operation"},
	)
)

func init() {
	prometheus.MustRegister(mqProduced, mqConsumed, mqDuration)
}

// Producer wraps kafka.Writer
type Producer struct {
	writer    *kafka.Writer
	dlqWriter *kafka.Writer // Dead Letter Queue writer
	logger    *logging.Logger
	tracer    trace.Tracer
}

// NewProducer creates a new Kafka producer.
func NewProducer(cfg config.KafkaConfig, logger *logging.Logger) *Producer {
	w := &kafka.Writer{
		Addr:         kafka.TCP(cfg.Brokers...),
		Topic:        cfg.Topic,
		Balancer:     &kafka.LeastBytes{},
		WriteTimeout: cfg.WriteTimeout,
		ReadTimeout:  cfg.ReadTimeout,
		BatchSize:    cfg.MaxBytes,
		BatchTimeout: 10 * time.Millisecond,
		Async:        cfg.Async,
		RequiredAcks: kafka.RequiredAcks(cfg.RequiredAcks),
	}

	var dlqWriter *kafka.Writer
	// Simple convention: DLQ topic is original topic + "-dlq"
	// In a real system, this might be configurable
	dlqWriter = &kafka.Writer{
		Addr:         kafka.TCP(cfg.Brokers...),
		Topic:        cfg.Topic + "-dlq",
		Balancer:     &kafka.LeastBytes{},
		WriteTimeout: cfg.WriteTimeout,
		ReadTimeout:  cfg.ReadTimeout,
		Async:        cfg.Async,
		RequiredAcks: kafka.RequiredAcks(cfg.RequiredAcks),
	}

	return &Producer{
		writer:    w,
		dlqWriter: dlqWriter,
		logger:    logger,
		tracer:    otel.Tracer("kafka-producer"),
	}
}

// Publish sends a message to Kafka.
func (p *Producer) Publish(ctx context.Context, key, value []byte) error {
	start := time.Now()
	ctx, span := p.tracer.Start(ctx, "Publish")
	defer span.End()

	defer func() {
		mqDuration.WithLabelValues(p.writer.Topic, "publish").Observe(time.Since(start).Seconds())
	}()

	// Inject trace context into headers
	headers := make([]kafka.Header, 0)
	carrier := propagation.MapCarrier{}
	otel.GetTextMapPropagator().Inject(ctx, carrier)
	for k, v := range carrier {
		headers = append(headers, kafka.Header{Key: k, Value: []byte(v)})
	}

	msg := kafka.Message{
		Key:     key,
		Value:   value,
		Time:    time.Now(),
		Headers: headers,
	}

	// Retry logic with exponential backoff
	var err error
	for i := 0; i < 3; i++ {
		err = p.writer.WriteMessages(ctx, msg)
		if err == nil {
			mqProduced.WithLabelValues(p.writer.Topic, "success").Inc()
			return nil
		}
		time.Sleep(time.Duration(1<<i) * 100 * time.Millisecond)
	}

	// If failed after retries, send to DLQ
	p.logger.ErrorContext(ctx, "failed to publish message, sending to DLQ", "error", err)
	mqProduced.WithLabelValues(p.writer.Topic, "failed").Inc()

	if dlqErr := p.dlqWriter.WriteMessages(ctx, msg); dlqErr != nil {
		p.logger.ErrorContext(ctx, "failed to publish to DLQ", "error", dlqErr)
		return fmt.Errorf("failed to publish to DLQ: %w (original error: %v)", dlqErr, err)
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
	tracer trace.Tracer
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
		tracer: otel.Tracer("kafka-consumer"),
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

		// Extract trace context
		carrier := propagation.MapCarrier{}
		for _, h := range m.Headers {
			carrier[h.Key] = string(h.Value)
		}
		ctx = otel.GetTextMapPropagator().Extract(ctx, carrier)
		ctx, span := c.tracer.Start(ctx, "Consume")

		start := time.Now()
		// Process message
		err = handler(ctx, m)
		duration := time.Since(start).Seconds()
		mqDuration.WithLabelValues(c.reader.Config().Topic, "consume").Observe(duration)

		if err != nil {
			span.RecordError(err)
			span.End()
			c.logger.Error("handler failed", "error", err)
			mqConsumed.WithLabelValues(c.reader.Config().Topic, "failed").Inc()
			// Decide whether to commit or not based on error type
			// For now, we assume manual commit is needed if auto-commit is disabled,
			// but kafka-go Reader with CommitInterval > 0 does auto-commit.
			// If we used FetchMessage, we need to CommitMessages manually.
			continue
		}

		mqConsumed.WithLabelValues(c.reader.Config().Topic, "success").Inc()
		span.End()

		if err := c.reader.CommitMessages(ctx, m); err != nil {
			c.logger.Error("failed to commit message", "error", err)
		}
	}
}

// Close closes the consumer.
func (c *Consumer) Close() error {
	return c.reader.Close()
}
