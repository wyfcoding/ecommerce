package config

import (
	"fmt"
	"os"
	"time"

	"github.com/BurntSushi/toml"
	"gorm.io/gorm/logger"
)

// Config 是顶层配置结构体，可被服务嵌入
type Config struct {
	Server         ServerConfig         `toml:"server"`
	Data           DataConfig           `toml:"data"`
	Log            LogConfig            `toml:"log"`
	JWT            JWTConfig            `toml:"jwt"`
	Snowflake      SnowflakeConfig      `toml:"snowflake"`
	MessageQueue   MessageQueueConfig   `toml:"messagequeue"`
	Minio          MinioConfig          `toml:"minio"`
	Hadoop         HadoopConfig         `toml:"hadoop"`
	Tracing        TracingConfig        `toml:"tracing"`
	Metrics        MetricsConfig        `toml:"metrics"`
	RateLimit      RateLimitConfig      `toml:"ratelimit"`
	CircuitBreaker CircuitBreakerConfig `toml:"circuitbreaker"`
	Cache          CacheConfig          `toml:"cache"`
	Lock           LockConfig           `toml:"lock"`
	Services       ServicesConfig       `toml:"services"`
}

// ServerConfig 定义了服务器配置
type ServerConfig struct {
	Name        string `toml:"name"`
	Environment string `toml:"environment"`
	HTTP        struct {
		Addr         string        `toml:"addr"`
		Port         int           `toml:"port"`
		Timeout      time.Duration `toml:"timeout"`
		ReadTimeout  time.Duration `toml:"read_timeout"`
		WriteTimeout time.Duration `toml:"write_timeout"`
		IdleTimeout  time.Duration `toml:"idle_timeout"`
	} `toml:"http"`
	GRPC struct {
		Addr           string        `toml:"addr"`
		Port           int           `toml:"port"`
		Timeout        time.Duration `toml:"timeout"`
		MaxRecvMsgSize int           `toml:"max_recv_msg_size"`
		MaxSendMsgSize int           `toml:"max_send_msg_size"`
	} `toml:"grpc"`
}

// DataConfig 定义了数据相关的配置
type DataConfig struct {
	Database      DatabaseConfig      `toml:"database"`
	Redis         RedisConfig         `toml:"redis"`
	MongoDB       MongoDBConfig       `toml:"mongodb"`
	ClickHouse    ClickHouseConfig    `toml:"clickhouse"`
	Neo4j         Neo4jConfig         `toml:"neo4j"`
	Elasticsearch ElasticsearchConfig `toml:"elasticsearch"`
}

// DatabaseConfig 定义了MySQL数据库配置
type DatabaseConfig struct {
	Driver          string          `toml:"driver"`
	DSN             string          `toml:"dsn"`
	MaxIdleConns    int             `toml:"max_idle_conns"`
	MaxOpenConns    int             `toml:"max_open_conns"`
	ConnMaxLifetime time.Duration   `toml:"conn_max_lifetime"`
	LogLevel        logger.LogLevel `toml:"log_level"`
	SlowThreshold   time.Duration   `toml:"slow_threshold"`
}

// RedisConfig 定义了Redis配置
type RedisConfig struct {
	Addr         string        `toml:"addr"`
	Password     string        `toml:"password"`
	DB           int           `toml:"db"`
	ReadTimeout  time.Duration `toml:"read_timeout"`
	WriteTimeout time.Duration `toml:"write_timeout"`
	PoolSize     int           `toml:"pool_size"`
	MinIdleConns int           `toml:"min_idle_conns"`
}

// MongoDBConfig 定义了MongoDB配置
type MongoDBConfig struct {
	URI            string        `toml:"uri"`
	Database       string        `toml:"database"`
	ConnectTimeout time.Duration `toml:"connect_timeout"`
	MinPoolSize    uint64        `toml:"min_pool_size"`
	MaxPoolSize    uint64        `toml:"max_pool_size"`
}

// ClickHouseConfig 定义了ClickHouse配置
type ClickHouseConfig struct {
	DSN             string        `toml:"dsn"`
	Database        string        `toml:"database"`
	Username        string        `toml:"username"`
	Password        string        `toml:"password"`
	DialTimeout     time.Duration `toml:"dial_timeout"`
	MaxOpenConns    int           `toml:"max_open_conns"`
	MaxIdleConns    int           `toml:"max_idle_conns"`
	ConnMaxLifetime time.Duration `toml:"conn_max_lifetime"`
}

// Neo4jConfig 定义了Neo4j配置
type Neo4jConfig struct {
	URI      string `toml:"uri"`
	Username string `toml:"username"`
	Password string `toml:"password"`
}

// ElasticsearchConfig 定义了Elasticsearch配置
type ElasticsearchConfig struct {
	Addresses  []string `toml:"addresses"`
	Username   string   `toml:"username"`
	Password   string   `toml:"password"`
	MaxRetries int      `toml:"max_retries"`
}

// LogConfig 定义了日志配置
type LogConfig struct {
	Level  string `toml:"level"`
	Format string `toml:"format"`
	Output string `toml:"output"`
}

// JWTConfig 定义了JWT配置
type JWTConfig struct {
	Secret string        `toml:"secret"`
	Issuer string        `toml:"issuer"`
	Expire time.Duration `toml:"expire_duration"`
}

// SnowflakeConfig 定义了Snowflake配置
type SnowflakeConfig struct {
	StartTime string `toml:"start_time"`
	MachineID int64  `toml:"machine_id"`
}

// MessageQueueConfig 定义了消息队列配置
type MessageQueueConfig struct {
	Kafka KafkaConfig `toml:"kafka"`
}

// KafkaConfig 定义了Kafka配置
type KafkaConfig struct {
	Brokers        []string      `toml:"brokers"`
	Topic          string        `toml:"topic"`
	GroupID        string        `toml:"group_id"`
	DialTimeout    time.Duration `toml:"dial_timeout"`
	ReadTimeout    time.Duration `toml:"read_timeout"`
	WriteTimeout   time.Duration `toml:"write_timeout"`
	MinBytes       int           `toml:"min_bytes"`
	MaxBytes       int           `toml:"max_bytes"`
	MaxWait        time.Duration `toml:"max_wait"`
	MaxAttempts    int           `toml:"max_attempts"`
	CommitInterval time.Duration `toml:"commit_interval"`
	RequiredAcks   int           `toml:"required_acks"`
	Async          bool          `toml:"async"`
}

// MinioConfig 定义了MinIO配置
type MinioConfig struct {
	Endpoint        string `toml:"endpoint"`
	AccessKeyID     string `toml:"access_key_id"`
	SecretAccessKey string `toml:"secret_access_key"`
	UseSSL          bool   `toml:"use_ssl"`
}

// HadoopConfig 定义了Hadoop配置
type HadoopConfig struct {
	HDFS HDFSConfig `toml:"hdfs"`
}

// HDFSConfig 定义了HDFS配置
type HDFSConfig struct {
	Addresses []string `toml:"addresses"`
	User      string   `toml:"user"`
}

// TracingConfig 定义了链路追踪配置
type TracingConfig struct {
	ServiceName    string `toml:"service_name"`
	JaegerEndpoint string `toml:"jaeger_endpoint"`
}

// MetricsConfig 定义了指标监控配置
type MetricsConfig struct {
	Enabled bool   `toml:"enabled"`
	Port    string `toml:"port"`
	Path    string `toml:"path"`
}

// RateLimitConfig 定义了限流配置
type RateLimitConfig struct {
	Enabled bool `toml:"enabled"`
	Rate    int  `toml:"rate"`
	Burst   int  `toml:"burst"`
}

// CircuitBreakerConfig 定义了熔断器配置
type CircuitBreakerConfig struct {
	Enabled     bool          `toml:"enabled"`
	MaxRequests uint32        `toml:"max_requests"`
	Interval    time.Duration `toml:"interval"`
	Timeout     time.Duration `toml:"timeout"`
}

// CacheConfig 定义了缓存配置
type CacheConfig struct {
	Prefix            string        `toml:"prefix"`
	DefaultExpiration time.Duration `toml:"default_expiration"`
	CleanupInterval   time.Duration `toml:"cleanup_interval"`
}

// LockConfig 定义了分布式锁配置
type LockConfig struct {
	Prefix            string        `toml:"prefix"`
	DefaultExpiration time.Duration `toml:"default_expiration"`
	MaxRetries        int           `toml:"max_retries"`
	RetryDelay        time.Duration `toml:"retry_delay"`
}

// ServicesConfig 定义了微服务地址配置
type ServicesConfig struct {
	User           ServiceAddr `toml:"user"`
	Product        ServiceAddr `toml:"product"`
	Order          ServiceAddr `toml:"order"`
	Cart           ServiceAddr `toml:"cart"`
	Payment        ServiceAddr `toml:"payment"`
	Inventory      ServiceAddr `toml:"inventory"`
	Marketing      ServiceAddr `toml:"marketing"`
	Notification   ServiceAddr `toml:"notification"`
	Search         ServiceAddr `toml:"search"`
	Recommendation ServiceAddr `toml:"recommendation"`
	Gateway        ServiceAddr `toml:"gateway"`
}

// ServiceAddr 定义了单个服务的地址
type ServiceAddr struct {
	GRPCAddr string `toml:"grpc_addr"`
	HTTPAddr string `toml:"http_addr"`
}

// Load 从 TOML 文件加载配置
func Load(path string, v interface{}) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}
	if err := toml.Unmarshal(data, v); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}
	return nil
}

// LoadFromEnv 从环境变量指定的路径加载配置
func LoadFromEnv(envKey string, v interface{}) error {
	path := os.Getenv(envKey)
	if path == "" {
		path = "configs/config.toml" // 默认路径
	}
	return Load(path, v)
}
