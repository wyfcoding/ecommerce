package kafka

import (
	"time"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

// Config 结构体用于 Kafka 客户端配置。
type Config struct {
	Brokers      []string      `toml:"brokers"`
	Topic        string        `toml:"topic"`
	GroupID      string        `toml:"group_id"`
	DialTimeout  time.Duration `toml:"dial_timeout"`
	ReadTimeout  time.Duration `toml:"read_timeout"`
	WriteTimeout time.Duration `toml:"write_timeout"`
}

// NewKafkaProducer 创建一个新的 Kafka 生产者实例。
func NewKafkaProducer(conf *Config) (*kafka.Writer, error) {
	writer := &kafka.Writer{
		Addr:     kafka.TCP(conf.Brokers...),
		Topic:    conf.Topic,
		Balancer: &kafka.LeastBytes{},
		Dialer: &kafka.Dialer{
			Timeout: conf.DialTimeout,
		},
		ReadTimeout:  conf.ReadTimeout,
		WriteTimeout: conf.WriteTimeout,
	}

	zap.S().Infof("Kafka producer created for brokers: %v, topic: %s", conf.Brokers, conf.Topic)

	return writer, nil
}

// NewKafkaConsumer 创建一个新的 Kafka 消费者实例。
func NewKafkaConsumer(conf *Config) (*kafka.Reader, error) {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     conf.Brokers,
		GroupID:     conf.GroupID,
		Topic:       conf.Topic,
		MinBytes:    10e3, // 10KB
		MaxBytes:    10e6, // 10MB
		MaxAttempts: 3,
		Dialer: &kafka.Dialer{
			Timeout: conf.DialTimeout,
		},
		ReadTimeout:  conf.ReadTimeout,
		WriteTimeout: conf.WriteTimeout,
	})

	zap.S().Infof("Kafka consumer connected to brokers: %v, topic: %s, group: %s", conf.Brokers, conf.Topic, conf.GroupID)

	return reader, nil
}
