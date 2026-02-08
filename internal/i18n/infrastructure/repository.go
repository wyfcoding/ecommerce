package infrastructure

import (
	"context"

	"github.com/wyfcoding/ecommerce/internal/i18n/domain"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type I18nRepository struct {
	db *gorm.DB
}

func NewI18nRepository(db *gorm.DB) *I18nRepository {
	return &I18nRepository{db: db}
}

func (r *I18nRepository) SaveLanguage(ctx context.Context, lang *domain.Language) error {
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "code"}},
		DoUpdates: clause.AssignmentColumns([]string{"name", "direction", "enabled"}),
	}).Create(lang).Error
}

func (r *I18nRepository) ListLanguages(ctx context.Context) ([]domain.Language, error) {
	var langs []domain.Language
	if err := r.db.WithContext(ctx).Where("enabled = ?", true).Find(&langs).Error; err != nil {
		return nil, err
	}
	return langs, nil
}

func (r *I18nRepository) SaveTranslation(ctx context.Context, trans *domain.Translation) error {
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "lang_code"}, {Name: "key"}},
		DoUpdates: clause.AssignmentColumns([]string{"value", "context", "namespace"}),
	}).Create(trans).Error
}

func (r *I18nRepository) GetTranslation(ctx context.Context, lang, key string) (*domain.Translation, error) {
	var t domain.Translation
	if err := r.db.WithContext(ctx).Where("lang_code = ? AND `key` = ?", lang, key).First(&t).Error; err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *I18nRepository) ListTranslations(ctx context.Context, lang string, keys []string, ns string) ([]domain.Translation, error) {
	var ts []domain.Translation
	tx := r.db.WithContext(ctx).Where("lang_code = ?", lang)

	if len(keys) > 0 {
		tx = tx.Where("`key` IN ?", keys)
	}
	if ns != "" {
		tx = tx.Where("namespace = ?", ns)
	}

	if err := tx.Find(&ts).Error; err != nil {
		return nil, err
	}
	return ts, nil
}
