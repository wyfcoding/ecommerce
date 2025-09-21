package minio

import (
	"context"
	"fmt"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"go.uber.org/zap"
)

// Config 结构体用于 MinIO 客户端配置。
type Config struct {
	Endpoint        string `toml:"endpoint"`
	AccessKeyID     string `toml:"access_key_id"`
	SecretAccessKey string `toml:"secret_access_key"`
	UseSSL          bool   `toml:"use_ssl"`
}

// NewMinioClient 创建一个新的 MinIO 客户端实例。
func NewMinioClient(conf *Config) (*minio.Client, error) {
	minioClient, err := minio.New(conf.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(conf.AccessKeyID, conf.SecretAccessKey, ""),
		Secure: conf.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create minio client: %w", err)
	}

	// 检查连接
	_, err = minioClient.ListBuckets(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to list minio buckets: %w", err)
	}

	zap.S().Infof("Successfully connected to MinIO at %s", conf.Endpoint)

	return minioClient, nil
}
