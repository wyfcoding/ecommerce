package application

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/wyfcoding/ecommerce/internal/file/domain"
)

// FileManager 处理文件管理的写操作。
type FileManager struct {
	repo   domain.FileRepository
	logger *slog.Logger
}

// NewFileManager creates a new FileManager instance.
func NewFileManager(repo domain.FileRepository, logger *slog.Logger) *FileManager {
	return &FileManager{
		repo:   repo,
		logger: logger,
	}
}

// UploadFile 模拟文件上传操作。
func (m *FileManager) UploadFile(ctx context.Context, name string, size int64, fileType domain.FileType, content []byte) (*domain.FileMetadata, error) {
	bucket := "default-bucket"
	path := fmt.Sprintf("/%s/%d/%s", bucket, time.Now().Unix(), name)
	url := fmt.Sprintf("http://localhost:9000%s", path)
	checksum := "simulated-checksum"

	file := domain.NewFileMetadata(name, size, fileType, path, url, checksum, bucket)

	if err := m.repo.Save(ctx, file); err != nil {
		m.logger.ErrorContext(ctx, "failed to save file metadata", "name", name, "error", err)
		return nil, err
	}

	m.logger.InfoContext(ctx, "file uploaded successfully (simulated)", "path", path, "file_id", file.ID)
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
