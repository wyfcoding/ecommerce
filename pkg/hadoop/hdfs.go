package hadoop

import (
	"context"
	"fmt"
	"io"

	"github.com/colinmarc/hdfs/v2"
	"go.uber.org/zap"
)

// Config 结构体用于 HDFS 客户端配置。
type Config struct {
	Addresses []string `toml:"addresses"`
	User      string   `toml:"user"`
}

// NewHDFSClient 创建一个新的 HDFS 客户端实例。
func NewHDFSClient(conf *Config) (*hdfs.Client, func(), error) {
	client, err := hdfs.NewClient(hdfs.ClientOptions{
		Addresses: conf.Addresses,
		User:      conf.User,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create HDFS client: %w", err)
	}

	// 检查连接 (例如，尝试列出根目录)
	_, err = client.ReadDir("/")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to HDFS: %w", err)
	}

	zap.S().Infof("Successfully connected to HDFS at %v", conf.Addresses)

	cleanup := func() {
		if client != nil {
			zap.S().Info("closing HDFS client...")
			if err := client.Close(); err != nil {
				zap.S().Errorf("failed to close hdfs client: %v", err)
			}
		}
	}

	return client, cleanup, nil
}

// WriteFileToHDFS 写入文件到 HDFS。
func WriteFileToHDFS(ctx context.Context, client *hdfs.Client, path string, reader io.Reader) error {
	writer, err := client.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create HDFS file %s: %w", path, err)
	}
	defer writer.Close()

	_, err = io.Copy(writer, reader)
	if err != nil {
		return fmt.Errorf("failed to write data to HDFS file %s: %w", path, err)
	}

	return nil
}

// ReadFileFromHDFS 从 HDFS 读取文件。
func ReadFileFromHDFS(ctx context.Context, client *hdfs.Client, path string) ([]byte, error) {
	reader, err := client.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open HDFS file %s: %w", path, err)
	}
	defer reader.Close()

	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read data from HDFS file %s: %w", path, err)
	}

	return data, nil
}
