package kafka

import (
	"time"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

// Config 结构体用于 Kafka 客户端配置, 增加了详细的生产者和消费者参数。
type Config struct {
	Brokers      []string      `toml:"brokers"`
	Topic        string        `toml:"topic"`
	GroupID      string        `toml:"group_id"`
	DialTimeout  time.Duration `toml:"dial_timeout"`
	ReadTimeout  time.Duration `toml:"read_timeout"`
	WriteTimeout time.Duration `toml:"write_timeout"`

	// Producer-specific options
	RequiredAcks int  `toml:"required_acks"`
	Async        bool `toml:"async"`

	// Consumer-specific options
	MinBytes       int           `toml:"min_bytes"`
	MaxBytes       int           `toml:"max_bytes"`
	MaxWait        time.Duration `toml:"max_wait"`
	MaxAttempts    int           `toml:"max_attempts"`
	CommitInterval time.Duration `toml:"commit_interval"`
}

// NewKafkaProducer 创建一个新的 Kafka 生产者实例。
// 增加了对 RequiredAcks 和 Async 的支持，并为Acks提供了安全默认值。
func NewKafkaProducer(conf *Config) (*kafka.Writer, func(), error) {
	// 如果未在配置中指定，默认为 kafka.RequireAll，确保最高的数据一致性。
	if conf.RequiredAcks == 0 {
		conf.RequiredAcks = int(kafka.RequireAll)
	}

	writer := &kafka.Writer{
		Addr:         kafka.TCP(conf.Brokers...),
		Topic:        conf.Topic,
		Balancer:     &kafka.LeastBytes{},
		RequiredAcks: kafka.RequiredAcks(conf.RequiredAcks),
		Async:        conf.Async,
		ReadTimeout:  conf.ReadTimeout,
		WriteTimeout: conf.WriteTimeout,
	}

	cleanup := func() {
		if err := writer.Close(); err != nil {
			zap.S().Errorf("failed to close kafka writer: %v", err)
		}
	}

	zap.S().Infof("Kafka producer created for brokers: %v, topic: %s, acks: %d, async: %t", conf.Brokers, conf.Topic, conf.RequiredAcks, conf.Async)

	return writer, cleanup, nil
}

// NewKafkaConsumer 创建一个新的 Kafka 消费者实例。
// 使用配置驱动的参数替换了硬编码值，并为关键参数提供了合理的默认值。
func NewKafkaConsumer(conf *Config) (*kafka.Reader, func(), error) {
	// 如果未配置，则提供合理的默认值。
	if conf.MinBytes == 0 {
		conf.MinBytes = 10e3 // 10KB
	}
	if conf.MaxBytes == 0 {
		conf.MaxBytes = 10e6 // 10MB
	}
	if conf.MaxAttempts == 0 {
		conf.MaxAttempts = 3
	}

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        conf.Brokers,
		GroupID:        conf.GroupID,
		Topic:          conf.Topic,
		MinBytes:       conf.MinBytes,
		MaxBytes:       conf.MaxBytes,
		MaxWait:        conf.MaxWait,
		MaxAttempts:    conf.MaxAttempts,
		CommitInterval: conf.CommitInterval, // 值为 0 表示同步提交, > 0 表示异步提交的间隔。
		Dialer: &kafka.Dialer{
			Timeout: conf.DialTimeout,
		},
	})

	cleanup := func() {
		if err := reader.Close(); err != nil {
			zap.S().Errorf("failed to close kafka reader: %v", err)
		}
	}

	zap.S().Infof("Kafka consumer connected to brokers: %v, topic: %s, group: %s", conf.Brokers, conf.Topic, conf.GroupID)

	return reader, cleanup, nil
}
