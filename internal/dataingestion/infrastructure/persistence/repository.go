package persistence

import (
	"context"

	"github.com/wyfcoding/ecommerce/internal/dataingestion/domain"

	"gorm.io/gorm"
)

type dataIngestionRepository struct {
	db *gorm.DB
}

// NewDataIngestionRepository 创建并返回一个新的 dataIngestionRepository 实例。
func NewDataIngestionRepository(db *gorm.DB) domain.DataIngestionRepository {
	return &dataIngestionRepository{db: db}
}

// --- Source methods ---

func (r *dataIngestionRepository) SaveSource(ctx context.Context, source *domain.IngestionSource) error {
	return r.db.WithContext(ctx).Save(source).Error
}

func (r *dataIngestionRepository) GetSource(ctx context.Context, id uint64) (*domain.IngestionSource, error) {
	var source domain.IngestionSource
	if err := r.db.WithContext(ctx).First(&source, id).Error; err != nil {
		return nil, err
	}
	return &source, nil
}

func (r *dataIngestionRepository) ListSources(ctx context.Context, offset, limit int) ([]*domain.IngestionSource, int64, error) {
	var list []*domain.IngestionSource
	var total int64

	db := r.db.WithContext(ctx).Model(&domain.IngestionSource{})

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := db.Offset(offset).Limit(limit).Order("created_at desc").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

func (r *dataIngestionRepository) UpdateSource(ctx context.Context, source *domain.IngestionSource) error {
	return r.db.WithContext(ctx).Save(source).Error
}

func (r *dataIngestionRepository) DeleteSource(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&domain.IngestionSource{}, id).Error
}

// --- Job methods ---

func (r *dataIngestionRepository) SaveJob(ctx context.Context, job *domain.IngestionJob) error {
	return r.db.WithContext(ctx).Save(job).Error
}

func (r *dataIngestionRepository) GetJob(ctx context.Context, id uint64) (*domain.IngestionJob, error) {
	var job domain.IngestionJob
	if err := r.db.WithContext(ctx).First(&job, id).Error; err != nil {
		return nil, err
	}
	return &job, nil
}

func (r *dataIngestionRepository) ListJobs(ctx context.Context, sourceID uint64, offset, limit int) ([]*domain.IngestionJob, int64, error) {
	var list []*domain.IngestionJob
	var total int64

	db := r.db.WithContext(ctx).Model(&domain.IngestionJob{})
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

func (r *dataIngestionRepository) UpdateJob(ctx context.Context, job *domain.IngestionJob) error {
	return r.db.WithContext(ctx).Save(job).Error
}

// --- Event methods ---

func (r *dataIngestionRepository) SaveEvent(ctx context.Context, event *domain.IngestedEvent) error {
	return r.db.WithContext(ctx).Save(event).Error
}

func (r *dataIngestionRepository) SaveEvents(ctx context.Context, events []*domain.IngestedEvent) error {
	return r.db.WithContext(ctx).Create(events).Error
}
