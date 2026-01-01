package application

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"log/slog"
	"time"

	"github.com/wyfcoding/ecommerce/internal/file/domain"
	"github.com/wyfcoding/pkg/storage"
)

// FileManager 处理文件管理的写操作。
type FileManager struct {
	repo    domain.FileRepository
	storage storage.Storage
	logger  *slog.Logger
}

// NewFileManager creates a new FileManager instance.
func NewFileManager(repo domain.FileRepository, storage storage.Storage, logger *slog.Logger) *FileManager {
	return &FileManager{
		repo:    repo,
		storage: storage,
		logger:  logger,
	}
}

// UploadFile 执行文件上传。
func (m *FileManager) UploadFile(ctx context.Context, name string, size int64, fileType domain.FileType, content []byte) (*domain.FileMetadata, error) {
	bucket := "default-bucket"
	// 构建唯一路径防止重名
	path := fmt.Sprintf("%d/%s", time.Now().Unix(), name)

	// 1. 调用真实存储引擎上传
	reader := bytes.NewReader(content)
	if err := m.storage.Upload(ctx, path, reader, size, string(fileType)); err != nil {
		m.logger.ErrorContext(ctx, "failed to upload file to storage", "path", path, "error", err)
		return nil, fmt.Errorf("storage upload failed: %w", err)
	}

	// 2. 计算真实校验和 (SHA256)
	hash := sha256.Sum256(content)
	checksum := fmt.Sprintf("%x", hash)

	// 3. 获取访问地址 (假设存储实现支持生成 URL，这里根据业务逻辑拼接或通过 GetPresignedURL 获取)
	// 简单起见，这里假设是公开 Bucket 并通过配置拼接
	url := fmt.Sprintf("/%s/%s", bucket, path)

	file := domain.NewFileMetadata(name, size, fileType, path, url, checksum, bucket)

	if err := m.repo.Save(ctx, file); err != nil {
		m.logger.ErrorContext(ctx, "failed to save file metadata", "name", name, "error", err)
		// 尝试从存储回滚已上传文件
		_ = m.storage.Delete(ctx, path)
		return nil, err
	}

	m.logger.InfoContext(ctx, "file uploaded successfully", "path", path, "file_id", file.ID)
	return file, nil
}

// DeleteFile 删除指定ID的文件。
func (m *FileManager) DeleteFile(ctx context.Context, id uint64) error {
	if err := m.repo.Delete(ctx, id); err != nil {
		m.logger.ErrorContext(ctx, "failed to delete file metadata", "file_id", id, "error", err)
		return err
	}
	return nil
}
