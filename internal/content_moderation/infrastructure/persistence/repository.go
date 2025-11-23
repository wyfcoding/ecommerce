package persistence

import (
	"context"
	"ecommerce/internal/content_moderation/domain/entity"
	"ecommerce/internal/content_moderation/domain/repository"

	"gorm.io/gorm"
)

type moderationRepository struct {
	db *gorm.DB
}

func NewModerationRepository(db *gorm.DB) repository.ModerationRepository {
	return &moderationRepository{db: db}
}

// ModerationRecord methods
func (r *moderationRepository) CreateRecord(ctx context.Context, record *entity.ModerationRecord) error {
	return r.db.WithContext(ctx).Create(record).Error
}

func (r *moderationRepository) GetRecord(ctx context.Context, id uint64) (*entity.ModerationRecord, error) {
	var record entity.ModerationRecord
	if err := r.db.WithContext(ctx).First(&record, id).Error; err != nil {
		return nil, err
	}
	return &record, nil
}

func (r *moderationRepository) UpdateRecord(ctx context.Context, record *entity.ModerationRecord) error {
	return r.db.WithContext(ctx).Save(record).Error
}

func (r *moderationRepository) ListRecords(ctx context.Context, status entity.ModerationStatus, offset, limit int) ([]*entity.ModerationRecord, int64, error) {
	var list []*entity.ModerationRecord
	var total int64

	db := r.db.WithContext(ctx).Model(&entity.ModerationRecord{})
	// If status is not -1 (assuming -1 means all), filter by status
	// But ModerationStatus is int8, so we can't use -1 easily if it's not defined.
	// Let's assume caller handles logic or we define a way to query all.
	// For now, let's query all if status is not passed (but status is value type).
	// Let's just filter by status for now.
	db = db.Where("status = ?", status)

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := db.Offset(offset).Limit(limit).Order("created_at desc").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// SensitiveWord methods
func (r *moderationRepository) CreateWord(ctx context.Context, word *entity.SensitiveWord) error {
	return r.db.WithContext(ctx).Create(word).Error
}

func (r *moderationRepository) GetWord(ctx context.Context, id uint64) (*entity.SensitiveWord, error) {
	var word entity.SensitiveWord
	if err := r.db.WithContext(ctx).First(&word, id).Error; err != nil {
		return nil, err
	}
	return &word, nil
}

func (r *moderationRepository) UpdateWord(ctx context.Context, word *entity.SensitiveWord) error {
	return r.db.WithContext(ctx).Save(word).Error
}

func (r *moderationRepository) DeleteWord(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&entity.SensitiveWord{}, id).Error
}

func (r *moderationRepository) ListWords(ctx context.Context, offset, limit int) ([]*entity.SensitiveWord, int64, error) {
	var list []*entity.SensitiveWord
	var total int64

	db := r.db.WithContext(ctx).Model(&entity.SensitiveWord{})

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := db.Offset(offset).Limit(limit).Order("created_at desc").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

func (r *moderationRepository) FindWord(ctx context.Context, word string) (*entity.SensitiveWord, error) {
	var w entity.SensitiveWord
	if err := r.db.WithContext(ctx).Where("word = ?", word).First(&w).Error; err != nil {
		return nil, err
	}
	return &w, nil
}
