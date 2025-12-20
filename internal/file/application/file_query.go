package application

import (
	"context"

	"github.com/wyfcoding/ecommerce/internal/file/domain"
)

// FileQuery handles read operations for file management.
type FileQuery struct {
	repo domain.FileRepository
}

// NewFileQuery creates a new FileQuery instance.
func NewFileQuery(repo domain.FileRepository) *FileQuery {
	return &FileQuery{
		repo: repo,
	}
}

// GetFile 获取指定ID的文件元数据信息。
func (q *FileQuery) GetFile(ctx context.Context, id uint64) (*domain.FileMetadata, error) {
	return q.repo.Get(ctx, id)
}

// ListFiles 获取文件元数据列表。
func (q *FileQuery) ListFiles(ctx context.Context, page, pageSize int) ([]*domain.FileMetadata, int64, error) {
	offset := (page - 1) * pageSize
	return q.repo.List(ctx, offset, pageSize)
}
