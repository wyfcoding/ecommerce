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

// NewFileService 创建文件服务门面实例。
func NewFileService(manager *FileManager, query *FileQuery) *FileService {
	return &FileService{
		manager: manager,
		query:   query,
	}
}

// --- 写操作（委托给 Manager）---

// UploadFile 上传文件并保存元数据。
func (s *FileService) UploadFile(ctx context.Context, name string, size int64, fileType domain.FileType, content []byte) (*domain.FileMetadata, error) {
	return s.manager.UploadFile(ctx, name, size, fileType, content)
}

// DeleteFile 删除指定ID的文件及其元数据。
func (s *FileService) DeleteFile(ctx context.Context, id uint64) error {
	return s.manager.DeleteFile(ctx, id)
}

// --- 读操作（委托给 Query）---

// GetFile 获取指定ID的文件元数据。
func (s *FileService) GetFile(ctx context.Context, id uint64) (*domain.FileMetadata, error) {
	return s.query.GetFile(ctx, id)
}

// ListFiles 获取文件元数据列表（分页）。
func (s *FileService) ListFiles(ctx context.Context, page, pageSize int) ([]*domain.FileMetadata, int64, error) {
	return s.query.ListFiles(ctx, page, pageSize)
}
