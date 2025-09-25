package config

import (
	"os"
	"time"

	"github.com/BurntSushi/toml"
)

// ServerConfig 定义了所有服务通用的服务器配置
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

// LogConfig 定义了所有服务通用的日志配置
type LogConfig struct {
	Level  string `toml:"level"`
	Format string `toml:"format"`
	Output string `toml:"output"`
}

// LoadConfig 是一个通用函数，用于从指定路径加载并解析 TOML 配置文件。
// configPtr 应该是一个指向服务特定配置结构体的指针。
func LoadConfig(path string, configPtr interface{}) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	err = toml.Unmarshal(data, configPtr)
	if err != nil {
		return err
	}
	return nil
}
