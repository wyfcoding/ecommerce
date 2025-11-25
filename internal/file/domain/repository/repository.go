package repository

import (
	"context"
	"github.com/wyfcoding/ecommerce/internal/file/domain/entity"
)

// FileRepository 文件仓储接口
type FileRepository interface {
	Save(ctx context.Context, file *entity.FileMetadata) error
	Get(ctx context.Context, id uint64) (*entity.FileMetadata, error)
	Delete(ctx context.Context, id uint64) error
	List(ctx context.Context, offset, limit int) ([]*entity.FileMetadata, int64, error)
}
