package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"gorm.io/gorm/logger"
)

// Config is the top-level configuration struct.
type Config struct {
	Server         ServerConfig         `mapstructure:"server" toml:"server"`
	Data           DataConfig           `mapstructure:"data" toml:"data"`
	Log            LogConfig            `mapstructure:"log" toml:"log"`
	JWT            JWTConfig            `mapstructure:"jwt" toml:"jwt"`
	Snowflake      SnowflakeConfig      `mapstructure:"snowflake" toml:"snowflake"`
	MessageQueue   MessageQueueConfig   `mapstructure:"messagequeue" toml:"messagequeue"`
	Minio          MinioConfig          `mapstructure:"minio" toml:"minio"`
	Hadoop         HadoopConfig         `mapstructure:"hadoop" toml:"hadoop"`
	Tracing        TracingConfig        `mapstructure:"tracing" toml:"tracing"`
	Metrics        MetricsConfig        `mapstructure:"metrics" toml:"metrics"`
	RateLimit      RateLimitConfig      `mapstructure:"ratelimit" toml:"ratelimit"`
	CircuitBreaker CircuitBreakerConfig `mapstructure:"circuitbreaker" toml:"circuitbreaker"`
	Cache          CacheConfig          `mapstructure:"cache" toml:"cache"`
	Lock           LockConfig           `mapstructure:"lock" toml:"lock"`
	Services       ServicesConfig       `mapstructure:"services" toml:"services"`
}

// ServerConfig defines server configuration.
type ServerConfig struct {
	Name        string `mapstructure:"name" toml:"name"`
	Environment string `mapstructure:"environment" toml:"environment"`
	HTTP        struct {
		Addr         string        `mapstructure:"addr" toml:"addr"`
		Port         int           `mapstructure:"port" toml:"port"`
		Timeout      time.Duration `mapstructure:"timeout" toml:"timeout"`
		ReadTimeout  time.Duration `mapstructure:"read_timeout" toml:"read_timeout"`
		WriteTimeout time.Duration `mapstructure:"write_timeout" toml:"write_timeout"`
		IdleTimeout  time.Duration `mapstructure:"idle_timeout" toml:"idle_timeout"`
	} `mapstructure:"http" toml:"http"`
	GRPC struct {
		Addr           string        `mapstructure:"addr" toml:"addr"`
		Port           int           `mapstructure:"port" toml:"port"`
		Timeout        time.Duration `mapstructure:"timeout" toml:"timeout"`
		MaxRecvMsgSize int           `mapstructure:"max_recv_msg_size" toml:"max_recv_msg_size"`
		MaxSendMsgSize int           `mapstructure:"max_send_msg_size" toml:"max_send_msg_size"`
	} `mapstructure:"grpc" toml:"grpc"`
}

// DataConfig defines data related configuration.
type DataConfig struct {
	Database      DatabaseConfig      `mapstructure:"database" toml:"database"`
	Shards        []DatabaseConfig    `mapstructure:"shards" toml:"shards"`
	Redis         RedisConfig         `mapstructure:"redis" toml:"redis"`
	BigCache      BigCacheConfig      `mapstructure:"bigcache" toml:"bigcache"`
	MongoDB       MongoDBConfig       `mapstructure:"mongodb" toml:"mongodb"`
	ClickHouse    ClickHouseConfig    `mapstructure:"clickhouse" toml:"clickhouse"`
	Neo4j         Neo4jConfig         `mapstructure:"neo4j" toml:"neo4j"`
	Elasticsearch ElasticsearchConfig `mapstructure:"elasticsearch" toml:"elasticsearch"`
}

// DatabaseConfig defines MySQL configuration.
type DatabaseConfig struct {
	Driver          string          `mapstructure:"driver" toml:"driver"`
	DSN             string          `mapstructure:"dsn" toml:"dsn"`
	MaxIdleConns    int             `mapstructure:"max_idle_conns" toml:"max_idle_conns"`
	MaxOpenConns    int             `mapstructure:"max_open_conns" toml:"max_open_conns"`
	ConnMaxLifetime time.Duration   `mapstructure:"conn_max_lifetime" toml:"conn_max_lifetime"`
	LogLevel        logger.LogLevel `mapstructure:"log_level" toml:"log_level"`
	SlowThreshold   time.Duration   `mapstructure:"slow_threshold" toml:"slow_threshold"`
}

// RedisConfig defines Redis configuration.
type RedisConfig struct {
	Addr         string        `mapstructure:"addr" toml:"addr"`
	Password     string        `mapstructure:"password" toml:"password"`
	DB           int           `mapstructure:"db" toml:"db"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout" toml:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout" toml:"write_timeout"`
	PoolSize     int           `mapstructure:"pool_size" toml:"pool_size"`
	MinIdleConns int           `mapstructure:"min_idle_conns" toml:"min_idle_conns"`
}

type BigCacheConfig struct {
	TTL   time.Duration `mapstructure:"ttl" toml:"ttl"`
	MaxMB int           `mapstructure:"max_mb" toml:"max_mb"`
}

// MongoDBConfig defines MongoDB configuration.
type MongoDBConfig struct {
	URI            string        `mapstructure:"uri" toml:"uri"`
	Database       string        `mapstructure:"database" toml:"database"`
	ConnectTimeout time.Duration `mapstructure:"connect_timeout" toml:"connect_timeout"`
	MinPoolSize    uint64        `mapstructure:"min_pool_size" toml:"min_pool_size"`
	MaxPoolSize    uint64        `mapstructure:"max_pool_size" toml:"max_pool_size"`
}

// ClickHouseConfig defines ClickHouse configuration.
type ClickHouseConfig struct {
	DSN             string        `mapstructure:"dsn" toml:"dsn"`
	Database        string        `mapstructure:"database" toml:"database"`
	Username        string        `mapstructure:"username" toml:"username"`
	Password        string        `mapstructure:"password" toml:"password"`
	DialTimeout     time.Duration `mapstructure:"dial_timeout" toml:"dial_timeout"`
	MaxOpenConns    int           `mapstructure:"max_open_conns" toml:"max_open_conns"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns" toml:"max_idle_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime" toml:"conn_max_lifetime"`
}

// Neo4jConfig defines Neo4j configuration.
type Neo4jConfig struct {
	URI      string `mapstructure:"uri" toml:"uri"`
	Username string `mapstructure:"username" toml:"username"`
	Password string `mapstructure:"password" toml:"password"`
}

// ElasticsearchConfig defines Elasticsearch configuration.
type ElasticsearchConfig struct {
	Addresses  []string `mapstructure:"addresses" toml:"addresses"`
	Username   string   `mapstructure:"username" toml:"username"`
	Password   string   `mapstructure:"password" toml:"password"`
	MaxRetries int      `mapstructure:"max_retries" toml:"max_retries"`
}

// LogConfig defines logging configuration.
type LogConfig struct {
	Level      string `mapstructure:"level" toml:"level"`
	Format     string `mapstructure:"format" toml:"format"`
	Output     string `mapstructure:"output" toml:"output"`
	MaxSize    int    `mapstructure:"max_size" toml:"max_size"`
	MaxBackups int    `mapstructure:"max_backups" toml:"max_backups"`
	MaxAge     int    `mapstructure:"max_age" toml:"max_age"`
	Compress   bool   `mapstructure:"compress" toml:"compress"`
}

// JWTConfig defines JWT configuration.
type JWTConfig struct {
	Secret string        `mapstructure:"secret" toml:"secret"`
	Issuer string        `mapstructure:"issuer" toml:"issuer"`
	Expire time.Duration `mapstructure:"expire_duration" toml:"expire_duration"`
}

// SnowflakeConfig defines Snowflake configuration.
type SnowflakeConfig struct {
	StartTime string `mapstructure:"start_time" toml:"start_time"`
	MachineID int64  `mapstructure:"machine_id" toml:"machine_id"`
}

// MessageQueueConfig defines message queue configuration.
type MessageQueueConfig struct {
	Kafka KafkaConfig `mapstructure:"kafka" toml:"kafka"`
}

// KafkaConfig defines Kafka configuration.
type KafkaConfig struct {
	Brokers        []string      `mapstructure:"brokers" toml:"brokers"`
	Topic          string        `mapstructure:"topic" toml:"topic"`
	GroupID        string        `mapstructure:"group_id" toml:"group_id"`
	DialTimeout    time.Duration `mapstructure:"dial_timeout" toml:"dial_timeout"`
	ReadTimeout    time.Duration `mapstructure:"read_timeout" toml:"read_timeout"`
	WriteTimeout   time.Duration `mapstructure:"write_timeout" toml:"write_timeout"`
	MinBytes       int           `mapstructure:"min_bytes" toml:"min_bytes"`
	MaxBytes       int           `mapstructure:"max_bytes" toml:"max_bytes"`
	MaxWait        time.Duration `mapstructure:"max_wait" toml:"max_wait"`
	MaxAttempts    int           `mapstructure:"max_attempts" toml:"max_attempts"`
	CommitInterval time.Duration `mapstructure:"commit_interval" toml:"commit_interval"`
	RequiredAcks   int           `mapstructure:"required_acks" toml:"required_acks"`
	Async          bool          `mapstructure:"async" toml:"async"`
}

// MinioConfig defines MinIO configuration.
type MinioConfig struct {
	Endpoint        string `mapstructure:"endpoint" toml:"endpoint"`
	AccessKeyID     string `mapstructure:"access_key_id" toml:"access_key_id"`
	SecretAccessKey string `mapstructure:"secret_access_key" toml:"secret_access_key"`
	UseSSL          bool   `mapstructure:"use_ssl" toml:"use_ssl"`
}

// HadoopConfig defines Hadoop configuration.
type HadoopConfig struct {
	HDFS HDFSConfig `mapstructure:"hdfs" toml:"hdfs"`
}

// HDFSConfig defines HDFS configuration.
type HDFSConfig struct {
	Addresses []string `mapstructure:"addresses" toml:"addresses"`
	User      string   `mapstructure:"user" toml:"user"`
}

// TracingConfig defines tracing configuration.
type TracingConfig struct {
	ServiceName  string `mapstructure:"service_name" toml:"service_name"`
	OTLPEndpoint string `mapstructure:"otlp_endpoint" toml:"otlp_endpoint"`
}

// MetricsConfig defines metrics configuration.
type MetricsConfig struct {
	Enabled bool   `mapstructure:"enabled" toml:"enabled"`
	Port    string `mapstructure:"port" toml:"port"`
	Path    string `mapstructure:"path" toml:"path"`
}

// RateLimitConfig defines rate limiting configuration.
type RateLimitConfig struct {
	Enabled bool `mapstructure:"enabled" toml:"enabled"`
	Rate    int  `mapstructure:"rate" toml:"rate"`
	Burst   int  `mapstructure:"burst" toml:"burst"`
}

// CircuitBreakerConfig defines circuit breaker configuration.
type CircuitBreakerConfig struct {
	Enabled     bool          `mapstructure:"enabled" toml:"enabled"`
	MaxRequests uint32        `mapstructure:"max_requests" toml:"max_requests"`
	Interval    time.Duration `mapstructure:"interval" toml:"interval"`
	Timeout     time.Duration `mapstructure:"timeout" toml:"timeout"`
}

// CacheConfig defines cache configuration.
type CacheConfig struct {
	Prefix            string        `mapstructure:"prefix" toml:"prefix"`
	DefaultExpiration time.Duration `mapstructure:"default_expiration" toml:"default_expiration"`
	CleanupInterval   time.Duration `mapstructure:"cleanup_interval" toml:"cleanup_interval"`
}

// LockConfig defines distributed lock configuration.
type LockConfig struct {
	Prefix            string        `mapstructure:"prefix" toml:"prefix"`
	DefaultExpiration time.Duration `mapstructure:"default_expiration" toml:"default_expiration"`
	MaxRetries        int           `mapstructure:"max_retries" toml:"max_retries"`
	RetryDelay        time.Duration `mapstructure:"retry_delay" toml:"retry_delay"`
}

// ServicesConfig defines service addresses.
type ServicesConfig struct {
	User           ServiceAddr `mapstructure:"user" toml:"user"`
	Product        ServiceAddr `mapstructure:"product" toml:"product"`
	Order          ServiceAddr `mapstructure:"order" toml:"order"`
	Cart           ServiceAddr `mapstructure:"cart" toml:"cart"`
	Payment        ServiceAddr `mapstructure:"payment" toml:"payment"`
	Inventory      ServiceAddr `mapstructure:"inventory" toml:"inventory"`
	Marketing      ServiceAddr `mapstructure:"marketing" toml:"marketing"`
	Notification   ServiceAddr `mapstructure:"notification" toml:"notification"`
	Search         ServiceAddr `mapstructure:"search" toml:"search"`
	Recommendation ServiceAddr `mapstructure:"recommendation" toml:"recommendation"`
	Gateway        ServiceAddr `mapstructure:"gateway" toml:"gateway"`
}

// ServiceAddr defines a single service address.
type ServiceAddr struct {
	GRPCAddr string `mapstructure:"grpc_addr" toml:"grpc_addr"`
	HTTPAddr string `mapstructure:"http_addr" toml:"http_addr"`
}

// Load loads configuration from file and environment variables.
func Load(path string, v interface{}) error {
	viper.SetConfigFile(path)
	viper.SetConfigType("toml")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	if err := viper.Unmarshal(v); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return nil
}

// WatchConfig watches for config changes and reloads them.
func WatchConfig(v interface{}, onChange func()) {
	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		fmt.Printf("Config file changed: %s\n", e.Name)
		if err := viper.Unmarshal(v); err != nil {
			fmt.Printf("Failed to unmarshal config after change: %v\n", err)
			return
		}
		if onChange != nil {
			onChange()
		}
	})
}
