package minio

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/minio/minio-go/v7"                 // 导入MinIO Go客户端库。
	"github.com/minio/minio-go/v7/pkg/credentials" // 导入MinIO凭证处理包。
)

// Config 结构体用于 MinIO 客户端配置。
// 它包含了连接MinIO服务器所需的所有参数。
type Config struct {
	Endpoint        string `toml:"endpoint"`          // MinIO服务的API端点地址，例如 "play.min.io"。
	AccessKeyID     string `toml:"access_key_id"`     // MinIO访问密钥ID。
	SecretAccessKey string `toml:"secret_access_key"` // MinIO秘密访问密钥。
	UseSSL          bool   `toml:"use_ssl"`           // 是否使用SSL/TLS连接MinIO。
}

// NewMinioClient 创建一个新的 MinIO 客户端实例。
// 它根据提供的配置建立与MinIO服务器的连接，并返回 `*minio.Client` 实例和一个用于清理的空函数。
// conf: 包含MinIO连接参数的配置结构体。
func NewMinioClient(conf *Config) (*minio.Client, func(), error) {
	// 使用配置的Endpoint、AccessKeyID、SecretAccessKey和UseSSL来创建MinIO客户端。
	minioClient, err := minio.New(conf.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(conf.AccessKeyID, conf.SecretAccessKey, ""), // 使用静态凭证V4进行认证。
		Secure: conf.UseSSL,                                                         // 配置是否使用SSL。
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create minio client: %w", err)
	}

	// 检查与MinIO服务器的连接是否成功。
	// 尝试列出桶可以验证连接是否建立且认证信息是否正确。
	_, err = minioClient.ListBuckets(context.Background())
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list minio buckets: %w", err)
	}

	slog.Info("Successfully connected to MinIO", "endpoint", conf.Endpoint)

	// 返回一个空的清理函数。
	// MinIO客户端通常不需要显式关闭连接，因为底层HTTP客户端会管理连接池。
	// 提供一个空的清理函数是为了保持接口一致性。
	cleanup := func() {}

	return minioClient, cleanup, nil
}
