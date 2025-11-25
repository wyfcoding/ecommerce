package application

import (
	"context"
	"fmt"
	"github.com/wyfcoding/ecommerce/internal/file/domain/entity"
	"github.com/wyfcoding/ecommerce/internal/file/domain/repository"
	"time"

	"log/slog"
)

type FileService struct {
	repo   repository.FileRepository
	logger *slog.Logger
}

func NewFileService(repo repository.FileRepository, logger *slog.Logger) *FileService {
	return &FileService{
		repo:   repo,
		logger: logger,
	}
}

// UploadFile 上传文件 (模拟)
func (s *FileService) UploadFile(ctx context.Context, name string, size int64, fileType entity.FileType, content []byte) (*entity.FileMetadata, error) {
	// Simulate file storage path generation
	bucket := "default-bucket"
	path := fmt.Sprintf("/%s/%d/%s", bucket, time.Now().Unix(), name)
	url := fmt.Sprintf("http://localhost:9000%s", path)
	checksum := "simulated-checksum"

	file := entity.NewFileMetadata(name, size, fileType, path, url, checksum, bucket)

	if err := s.repo.Save(ctx, file); err != nil {
		s.logger.Error("failed to save file metadata", "error", err)
		return nil, err
	}

	s.logger.Info("File uploaded successfully (simulated)", "path", path)
	return file, nil
}

// GetFile 获取文件信息
func (s *FileService) GetFile(ctx context.Context, id uint64) (*entity.FileMetadata, error) {
	return s.repo.Get(ctx, id)
}

// DeleteFile 删除文件
func (s *FileService) DeleteFile(ctx context.Context, id uint64) error {
	// Simulate deletion from storage
	if err := s.repo.Delete(ctx, id); err != nil {
		s.logger.Error("failed to delete file metadata", "error", err)
		return err
	}
	return nil
}

// ListFiles 获取文件列表
func (s *FileService) ListFiles(ctx context.Context, page, pageSize int) ([]*entity.FileMetadata, int64, error) {
	offset := (page - 1) * pageSize
	return s.repo.List(ctx, offset, pageSize)
}
