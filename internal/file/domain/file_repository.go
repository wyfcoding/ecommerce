package domain

import (
	"context"
)

// FileRepository 是文件模块的仓储接口。
type FileRepository interface {
	Save(ctx context.Context, file *FileMetadata) error
	Get(ctx context.Context, id uint64) (*FileMetadata, error)
	Delete(ctx context.Context, id uint64) error
	List(ctx context.Context, offset, limit int) ([]*FileMetadata, int64, error)
}
