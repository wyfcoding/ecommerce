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

// UploadFile 处理文件上传全流程：物理存储写入 -> 元数据落库 -> 异常回滚。
func (m *FileManager) UploadFile(ctx context.Context, name string, size int64, fileType domain.FileType, content []byte) (*domain.FileMetadata, error) {
	bucket := "default-bucket"
	// 构建按时间分片的唯一存储路径，防止文件名冲突
	path := fmt.Sprintf("%d/%s", time.Now().Unix(), name)

	// 1. 调用通用存储接口执行物理上传
	reader := bytes.NewReader(content)
	if err := m.storage.Upload(ctx, path, reader, size, string(fileType)); err != nil {
		m.logger.ErrorContext(ctx, "failed to upload file to storage", "path", path, "error", err)
		return nil, fmt.Errorf("storage upload failed: %w", err)
	}

	// 2. 计算文件的 SHA256 校验和，用于数据完整性校验
	hash := sha256.Sum256(content)
	checksum := fmt.Sprintf("%x", hash)

	// 3. 构建访问 URL（此处为演示逻辑，实际应根据存储配置动态生成）
	url := fmt.Sprintf("/%s/%s", bucket, path)

	file := domain.NewFileMetadata(name, size, fileType, path, url, checksum, bucket)

	// 4. 持久化元数据。若失败，需同步清理已上传的物理文件以保证一致性
	if err := m.repo.Save(ctx, file); err != nil {
		m.logger.ErrorContext(ctx, "failed to save file metadata", "name", name, "error", err)
		if delErr := m.storage.Delete(ctx, path); delErr != nil {
			m.logger.ErrorContext(ctx, "failed to rollback storage file after DB failure", "path", path, "error", delErr)
		}
		return nil, err
	}

	m.logger.InfoContext(ctx, "file uploaded successfully", "path", path, "file_id", file.ID)
	return file, nil
}

// DeleteFile 彻底删除指定 ID 的文件，包含数据库记录与物理存储。
func (m *FileManager) DeleteFile(ctx context.Context, id uint64) error {
	// 1. 查找元数据获取存储路径
	file, err := m.repo.Get(ctx, id)
	if err != nil {
		m.logger.ErrorContext(ctx, "failed to get file metadata for deletion", "file_id", id, "error", err)
		return err
	}
	if file == nil {
		return nil // 幂等处理
	}

	// 2. 从数据库删除记录
	if err := m.repo.Delete(ctx, id); err != nil {
		m.logger.ErrorContext(ctx, "failed to delete file metadata", "file_id", id, "error", err)
		return err
	}

	// 3. 同步删除物理存储中的文件
	if err := m.storage.Delete(ctx, file.Path); err != nil {
		m.logger.WarnContext(ctx, "failed to delete physical file from storage", "path", file.Path, "error", err)
	}

	m.logger.InfoContext(ctx, "file deleted successfully", "file_id", id, "path", file.Path)
	return nil
}
