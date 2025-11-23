package persistence

import (
	"context"
	"ecommerce/internal/data_ingestion/domain/entity"
	"ecommerce/internal/data_ingestion/domain/repository"

	"gorm.io/gorm"
)

type dataIngestionRepository struct {
	db *gorm.DB
}

func NewDataIngestionRepository(db *gorm.DB) repository.DataIngestionRepository {
	return &dataIngestionRepository{db: db}
}

// Source methods
func (r *dataIngestionRepository) SaveSource(ctx context.Context, source *entity.IngestionSource) error {
	return r.db.WithContext(ctx).Save(source).Error
}

func (r *dataIngestionRepository) GetSource(ctx context.Context, id uint64) (*entity.IngestionSource, error) {
	var source entity.IngestionSource
	if err := r.db.WithContext(ctx).First(&source, id).Error; err != nil {
		return nil, err
	}
	return &source, nil
}

func (r *dataIngestionRepository) ListSources(ctx context.Context, offset, limit int) ([]*entity.IngestionSource, int64, error) {
	var list []*entity.IngestionSource
	var total int64

	db := r.db.WithContext(ctx).Model(&entity.IngestionSource{})

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := db.Offset(offset).Limit(limit).Order("created_at desc").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

func (r *dataIngestionRepository) UpdateSource(ctx context.Context, source *entity.IngestionSource) error {
	return r.db.WithContext(ctx).Save(source).Error
}

func (r *dataIngestionRepository) DeleteSource(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&entity.IngestionSource{}, id).Error
}

// Job methods
func (r *dataIngestionRepository) SaveJob(ctx context.Context, job *entity.IngestionJob) error {
	return r.db.WithContext(ctx).Save(job).Error
}

func (r *dataIngestionRepository) GetJob(ctx context.Context, id uint64) (*entity.IngestionJob, error) {
	var job entity.IngestionJob
	if err := r.db.WithContext(ctx).First(&job, id).Error; err != nil {
		return nil, err
	}
	return &job, nil
}

func (r *dataIngestionRepository) ListJobs(ctx context.Context, sourceID uint64, offset, limit int) ([]*entity.IngestionJob, int64, error) {
	var list []*entity.IngestionJob
	var total int64

	db := r.db.WithContext(ctx).Model(&entity.IngestionJob{})
	if sourceID != 0 {
		db = db.Where("source_id = ?", sourceID)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := db.Offset(offset).Limit(limit).Order("created_at desc").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

func (r *dataIngestionRepository) UpdateJob(ctx context.Context, job *entity.IngestionJob) error {
	return r.db.WithContext(ctx).Save(job).Error
}
