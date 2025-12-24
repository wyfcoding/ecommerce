package persistence

import (
	"context"

	"github.com/wyfcoding/ecommerce/internal/contentmoderation/domain"

	"gorm.io/gorm"
)

type moderationRepository struct {
	db *gorm.DB
}

// NewModerationRepository 创建并返回一个新的 moderationRepository 实例。
func NewModerationRepository(db *gorm.DB) domain.ModerationRepository {
	return &moderationRepository{db: db}
}

// --- ModerationRecord methods ---

func (r *moderationRepository) CreateRecord(ctx context.Context, record *domain.ModerationRecord) error {
	return r.db.WithContext(ctx).Create(record).Error
}

func (r *moderationRepository) GetRecord(ctx context.Context, id uint64) (*domain.ModerationRecord, error) {
	var record domain.ModerationRecord
	if err := r.db.WithContext(ctx).First(&record, id).Error; err != nil {
		return nil, err
	}
	return &record, nil
}

func (r *moderationRepository) UpdateRecord(ctx context.Context, record *domain.ModerationRecord) error {
	return r.db.WithContext(ctx).Save(record).Error
}

func (r *moderationRepository) ListRecords(ctx context.Context, status domain.ModerationStatus, offset, limit int) ([]*domain.ModerationRecord, int64, error) {
	var list []*domain.ModerationRecord
	var total int64

	db := r.db.WithContext(ctx).Model(&domain.ModerationRecord{})
	db = db.Where("status = ?", status)

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := db.Offset(offset).Limit(limit).Order("created_at desc").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// --- SensitiveWord methods ---

func (r *moderationRepository) CreateWord(ctx context.Context, word *domain.SensitiveWord) error {
	return r.db.WithContext(ctx).Create(word).Error
}

func (r *moderationRepository) GetWord(ctx context.Context, id uint64) (*domain.SensitiveWord, error) {
	var word domain.SensitiveWord
	if err := r.db.WithContext(ctx).First(&word, id).Error; err != nil {
		return nil, err
	}
	return &word, nil
}

func (r *moderationRepository) UpdateWord(ctx context.Context, word *domain.SensitiveWord) error {
	return r.db.WithContext(ctx).Save(word).Error
}

func (r *moderationRepository) DeleteWord(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&domain.SensitiveWord{}, id).Error
}

func (r *moderationRepository) ListWords(ctx context.Context, offset, limit int) ([]*domain.SensitiveWord, int64, error) {
	var list []*domain.SensitiveWord
	var total int64

	db := r.db.WithContext(ctx).Model(&domain.SensitiveWord{})

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := db.Offset(offset).Limit(limit).Order("created_at desc").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

func (r *moderationRepository) FindWord(ctx context.Context, word string) (*domain.SensitiveWord, error) {
	var w domain.SensitiveWord
	if err := r.db.WithContext(ctx).Where("word = ?", word).First(&w).Error; err != nil {
		return nil, err
	}
	return &w, nil
}
