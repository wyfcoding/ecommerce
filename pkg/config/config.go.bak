package config

import (
	"os"
	"time"

	"github.com/BurntSushi/toml"
)

// Config 是顶层配置结构体，可被服务嵌入。
type Config struct {
	Server    ServerConfig    `toml:"server"`
	Data      DataConfig      `toml:"data"`
	Log       LogConfig       `toml:"log"`
	JWT       JWTConfig       `toml:"jwt"`
	Snowflake SnowflakeConfig `toml:"snowflake"`
}

// ServerConfig 定义了服务器配置。
type ServerConfig struct {
	HTTP struct {
		Addr    string        `toml:"addr"`
		Port    int           `toml:"port"`
		Timeout time.Duration `toml:"timeout"`
	} `toml:"http"`
	GRPC struct {
		Addr    string        `toml:"addr"`
		Port    int           `toml:"port"`
		Timeout time.Duration `toml:"timeout"`
	} `toml:"grpc"`
}

// DataConfig 定义了数据相关的配置。
type DataConfig struct {
	Database DatabaseConfig `toml:"database"`
	Redis    RedisConfig    `toml:"redis"`
}

// DatabaseConfig 定义了数据库配置。
type DatabaseConfig struct {
	Driver string `toml:"driver"`
	DSN    string `toml:"dsn"`
}

// RedisConfig 定义了 Redis 配置。
type RedisConfig struct {
	Addr         string        `toml:"addr"`
	Password     string        `toml:"password"`
	DB           int           `toml:"db"`
	ReadTimeout  time.Duration `toml:"read_timeout"`
	WriteTimeout time.Duration `toml:"write_timeout"`
}

// LogConfig 定义了日志配置。
type LogConfig struct {
	Level  string `toml:"level"`
	Format string `toml:"format"`
	Output string `toml:"output"`
}

// JWTConfig 定义了 JWT 配置。
type JWTConfig struct {
	Secret string        `toml:"secret"`
	Issuer string        `toml:"issuer"`
	Expire time.Duration `toml:"expire_duration"`
}

// SnowflakeConfig 定义了 Snowflake 配置。
type SnowflakeConfig struct {
	StartTime string `toml:"start_time"`
	MachineID int64  `toml:"machine_id"`
}

// Load 从 TOML 文件加载配置。
func Load(path string, v interface{}) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return toml.Unmarshal(data, v)
}