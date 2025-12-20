package application

import (
	"context"

	"github.com/wyfcoding/ecommerce/internal/file/domain"
)

// FileService 作为文件管理操作的门面。
type FileService struct {
	manager *FileManager
	query   *FileQuery
}

// NewFileService creates a new FileService facade.
func NewFileService(manager *FileManager, query *FileQuery) *FileService {
	return &FileService{
		manager: manager,
		query:   query,
	}
}

// --- 写操作（委托给 Manager）---

func (s *FileService) UploadFile(ctx context.Context, name string, size int64, fileType domain.FileType, content []byte) (*domain.FileMetadata, error) {
	return s.manager.UploadFile(ctx, name, size, fileType, content)
}

func (s *FileService) DeleteFile(ctx context.Context, id uint64) error {
	return s.manager.DeleteFile(ctx, id)
}

// --- 读操作（委托给 Query）---

func (s *FileService) GetFile(ctx context.Context, id uint64) (*domain.FileMetadata, error) {
	return s.query.GetFile(ctx, id)
}

func (s *FileService) ListFiles(ctx context.Context, page, pageSize int) ([]*domain.FileMetadata, int64, error) {
	return s.query.ListFiles(ctx, page, pageSize)
}
