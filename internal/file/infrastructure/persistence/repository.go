package persistence

import (
	"context"
	"github.com/wyfcoding/ecommerce/internal/file/domain/entity"
	"github.com/wyfcoding/ecommerce/internal/file/domain/repository"

	"gorm.io/gorm"
)

type fileRepository struct {
	db *gorm.DB
}

func NewFileRepository(db *gorm.DB) repository.FileRepository {
	return &fileRepository{db: db}
}

func (r *fileRepository) Save(ctx context.Context, file *entity.FileMetadata) error {
	return r.db.WithContext(ctx).Save(file).Error
}

func (r *fileRepository) Get(ctx context.Context, id uint64) (*entity.FileMetadata, error) {
	var file entity.FileMetadata
	if err := r.db.WithContext(ctx).First(&file, id).Error; err != nil {
		return nil, err
	}
	return &file, nil
}

func (r *fileRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&entity.FileMetadata{}, id).Error
}

func (r *fileRepository) List(ctx context.Context, offset, limit int) ([]*entity.FileMetadata, int64, error) {
	var list []*entity.FileMetadata
	var total int64

	db := r.db.WithContext(ctx).Model(&entity.FileMetadata{})

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := db.Offset(offset).Limit(limit).Order("created_at desc").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}
