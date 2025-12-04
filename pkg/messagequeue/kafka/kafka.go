package kafka

import (
	"context"
	"fmt"
	"time"

	"github.com/wyfcoding/ecommerce/pkg/config"
	"github.com/wyfcoding/ecommerce/pkg/logging"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/segmentio/kafka-go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

var (
	// mqProduced 是一个Prometheus计数器，用于统计生产成功的消息总数。
	// 标签包括主题（topic）和发送状态（status）。
	mqProduced = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "mq_produced_total",
			Help: "The total number of messages produced",
		},
		[]string{"topic", "status"},
	)
	// mqConsumed 是一个Prometheus计数器，用于统计消费成功的消息总数。
	// 标签包括主题（topic）和消费状态（status）。
	mqConsumed = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "mq_consumed_total",
			Help: "The total number of messages consumed",
		},
		[]string{"topic", "status"},
	)
	// mqDuration 是一个Prometheus直方图，用于记录MQ操作（生产/消费）的耗时。
	// 标签包括主题（topic）和操作类型（operation）。
	mqDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "mq_operation_duration_seconds",
			Help:    "The duration of mq operations",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"topic", "operation"},
	)
)

// init 函数在包加载时自动执行，用于注册Prometheus指标。
func init() {
	prometheus.MustRegister(mqProduced, mqConsumed, mqDuration)
}

// Producer 封装了 `github.com/segmentio/kafka-go` 的 `kafka.Writer`，用于向Kafka生产消息。
// 它还包括一个死信队列（DLQ）写入器，用于处理发送失败的消息。
type Producer struct {
	writer    *kafka.Writer
	dlqWriter *kafka.Writer // Dead Letter Queue writer
	logger    *logging.Logger
	tracer    trace.Tracer
}

// NewProducer 创建一个新的Kafka生产者实例。
// cfg 参数提供了Kafka连接和Topic的配置信息。
// logger 用于记录生产者操作日志。
func NewProducer(cfg config.KafkaConfig, logger *logging.Logger) *Producer {
	w := &kafka.Writer{
		Addr:         kafka.TCP(cfg.Brokers...),
		Topic:        cfg.Topic,
		Balancer:     &kafka.LeastBytes{},                  // 负载均衡器，选择分区策略
		WriteTimeout: cfg.WriteTimeout,                     // 写入操作超时时间
		ReadTimeout:  cfg.ReadTimeout,                      // 读取操作超时时间
		BatchSize:    cfg.MaxBytes,                         // 批量发送的消息大小
		BatchTimeout: 10 * time.Millisecond,                // 批量发送的超时时间
		Async:        cfg.Async,                            // 是否启用异步发送
		RequiredAcks: kafka.RequiredAcks(cfg.RequiredAcks), // 生产者确认机制
	}

	// 约定死信队列的主题名称为原主题名加上 "-dlq" 后缀。
	// 在实际系统中，这通常是可配置的。
	dlqWriter := &kafka.Writer{
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
		tracer:    otel.Tracer("kafka-producer"), // OpenTelemetry 追踪器
	}
}

// Publish 发送消息到Kafka。
// 该方法集成了OpenTelemetry追踪上下文的注入、消息发送重试逻辑以及发送失败后的死信队列处理。
func (p *Producer) Publish(ctx context.Context, key, value []byte) error {
	start := time.Now()
	// 开始一个新的追踪span，名称为 "Publish"。
	ctx, span := p.tracer.Start(ctx, "Publish")
	defer span.End() // 确保span在函数退出时结束。

	defer func() {
		// 记录消息发布操作的耗时。
		mqDuration.WithLabelValues(p.writer.Topic, "publish").Observe(time.Since(start).Seconds())
	}()

	// 将当前追踪上下文注入到Kafka消息头中，以便消费者可以提取并继续追踪链。
	headers := make([]kafka.Header, 0)
	carrier := propagation.MapCarrier{}
	otel.GetTextMapPropagator().Inject(ctx, carrier) // 注入追踪上下文到MapCarrier
	for k, v := range carrier {
		headers = append(headers, kafka.Header{Key: k, Value: []byte(v)})
	}

	msg := kafka.Message{
		Key:     key,
		Value:   value,
		Time:    time.Now(),
		Headers: headers,
	}

	// 消息发送的重试逻辑，采用指数退避策略。
	var err error
	for i := 0; i < 3; i++ { // 最多重试3次
		err = p.writer.WriteMessages(ctx, msg)
		if err == nil {
			// 发送成功，记录成功指标并返回。
			mqProduced.WithLabelValues(p.writer.Topic, "success").Inc()
			return nil
		}
		// 重试前短暂休眠。
		time.Sleep(time.Duration(1<<i) * 100 * time.Millisecond)
	}

	// 如果所有重试都失败，则将消息发送到死信队列。
	p.logger.ErrorContext(ctx, "failed to publish message, sending to DLQ", "error", err)
	mqProduced.WithLabelValues(p.writer.Topic, "failed").Inc()

	if dlqErr := p.dlqWriter.WriteMessages(ctx, msg); dlqErr != nil {
		p.logger.ErrorContext(ctx, "failed to publish to DLQ", "error", dlqErr)
		// 如果发送到DLQ也失败，则返回一个包含两个错误的详细错误信息。
		return fmt.Errorf("failed to publish to DLQ: %w (original error: %v)", dlqErr, err)
	}

	return nil // 如果成功发送到DLQ，则返回nil，表示消息已妥善处理。
}

// Close 关闭生产者及其底层的Kafka writer。
func (p *Producer) Close() error {
	return p.writer.Close()
}

// Consumer 封装了 `github.com/segmentio/kafka-go` 的 `kafka.Reader`，用于从Kafka消费消息。
type Consumer struct {
	reader *kafka.Reader
	logger *logging.Logger
	tracer trace.Tracer
}

// NewConsumer 创建一个新的Kafka消费者实例。
// cfg 参数提供了Kafka连接、Topic和消费者组的配置信息。
// logger 用于记录消费者操作日志。
func NewConsumer(cfg config.KafkaConfig, logger *logging.Logger) *Consumer {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:         cfg.Brokers,
		GroupID:         cfg.GroupID,
		Topic:           cfg.Topic,
		MinBytes:        cfg.MinBytes,
		MaxBytes:        cfg.MaxBytes,
		MaxWait:         cfg.MaxWait,
		ReadLagInterval: -1,               // 不上报消费者组延迟指标
		CommitInterval:  0,                // 如果为0，表示手动提交，否则自动提交
		StartOffset:     kafka.LastOffset, // 消费者启动时从最新的偏移量开始消费
	})

	return &Consumer{
		reader: r,
		logger: logger,
		tracer: otel.Tracer("kafka-consumer"), // OpenTelemetry 追踪器
	}
}

// Consume 从Kafka读取消息并使用提供的handler处理它们。
// 这是一个阻塞方法，会持续消费消息直到上下文被取消或发生不可恢复的错误。
// 该方法集成了OpenTelemetry追踪上下文的提取和消费者指标的记录。
func (c *Consumer) Consume(ctx context.Context, handler func(ctx context.Context, msg kafka.Message) error) error {
	for {
		// FetchMessage 是一个阻塞调用，它会等待直到有新的消息可用，或者上下文被取消。
		m, err := c.reader.FetchMessage(ctx)
		if err != nil {
			// 如果是上下文取消错误，表示消费者应该停止。
			if ctx.Err() != nil {
				return ctx.Err()
			}
			// 其他错误，记录并继续尝试获取消息。
			c.logger.Error("failed to fetch message", "error", err)
			continue
		}

		// 从Kafka消息头中提取追踪上下文，以便继续追踪链。
		carrier := propagation.MapCarrier{}
		for _, h := range m.Headers {
			carrier[h.Key] = string(h.Value)
		}
		// 使用父上下文（即当前的ctx）作为基础，提取并创建一个新的上下文。
		ctx = otel.GetTextMapPropagator().Extract(ctx, carrier)
		// 为当前消息的处理启动一个新的追踪span。
		ctx, span := c.tracer.Start(ctx, "Consume")

		start := time.Now()
		// 调用业务处理函数。
		err = handler(ctx, m)
		duration := time.Since(start).Seconds()
		// 记录消息消费操作的耗时。
		mqDuration.WithLabelValues(c.reader.Config().Topic, "consume").Observe(duration)

		if err != nil {
			// 如果处理器返回错误，记录错误到span中。
			span.RecordError(err)
			span.End()
			c.logger.Error("handler failed", "error", err)
			// 记录消费失败指标。
			mqConsumed.WithLabelValues(c.reader.Config().Topic, "failed").Inc()
			// 注意：如果 CommitInterval > 0，kafka-go Reader 会自动提交。
			// 如果这里希望不提交，需要确保 CommitInterval 设置为 0 并手动控制提交。
			continue // 继续下一条消息，不提交当前消息的offset。
		}

		// 消息处理成功，记录成功指标。
		mqConsumed.WithLabelValues(c.reader.Config().Topic, "success").Inc()
		span.End()

		// 消息处理成功后，手动提交消息的offset。
		if err := c.reader.CommitMessages(ctx, m); err != nil {
			c.logger.Error("failed to commit message", "error", err)
		}
	}
}

// Close 关闭消费者及其底层的Kafka reader。
func (c *Consumer) Close() error {
	return c.reader.Close()
}
