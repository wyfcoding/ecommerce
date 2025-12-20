package persistence

import (
	"context"

	"github.com/wyfcoding/ecommerce/internal/file/domain"

	"gorm.io/gorm"
)

type fileRepository struct {
	db *gorm.DB
}

// NewFileRepository 创建并返回一个新的 fileRepository 实例。
func NewFileRepository(db *gorm.DB) domain.FileRepository {
	return &fileRepository{db: db}
}

func (r *fileRepository) Save(ctx context.Context, file *domain.FileMetadata) error {
	return r.db.WithContext(ctx).Save(file).Error
}

func (r *fileRepository) Get(ctx context.Context, id uint64) (*domain.FileMetadata, error) {
	var file domain.FileMetadata
	if err := r.db.WithContext(ctx).First(&file, id).Error; err != nil {
		return nil, err
	}
	return &file, nil
}

func (r *fileRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&domain.FileMetadata{}, id).Error
}

func (r *fileRepository) List(ctx context.Context, offset, limit int) ([]*domain.FileMetadata, int64, error) {
	var list []*domain.FileMetadata
	var total int64

	db := r.db.WithContext(ctx).Model(&domain.FileMetadata{})

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := db.Offset(offset).Limit(limit).Order("created_at desc").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}
