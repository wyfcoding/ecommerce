package application

import (
	"context"

	"github.com/wyfcoding/ecommerce/internal/i18n/domain"
)

type I18nService struct {
	repo domain.I18nRepository
}

func NewI18nService(repo domain.I18nRepository) *I18nService {
	return &I18nService{repo: repo}
}

func (s *I18nService) GetTranslation(ctx context.Context, lang, key string) (string, bool, error) {
	t, err := s.repo.GetTranslation(ctx, lang, key)
	if err != nil {
		return "", false, nil // return empty if not found, suppress error for business logic
	}
	return t.Value, true, nil
}

func (s *I18nService) PutTranslation(ctx context.Context, lang, key, value, contextStr, ns string) error {
	t := domain.NewTranslation(lang, key, value, contextStr, ns)
	return s.repo.SaveTranslation(ctx, t)
}

func (s *I18nService) ListTranslations(ctx context.Context, lang string, keys []string, ns string) (map[string]string, error) {
	list, err := s.repo.ListTranslations(ctx, lang, keys, ns)
	if err != nil {
		return nil, err
	}

	result := make(map[string]string)
	for _, item := range list {
		result[item.Key] = item.Value
	}
	return result, nil
}

func (s *I18nService) ListLanguages(ctx context.Context) ([]domain.Language, error) {
	return s.repo.ListLanguages(ctx)
}

func (s *I18nService) InitDefaults(ctx context.Context) error {
	// Initialize default languages
	defaults := []*domain.Language{
		domain.NewLanguage("en-US", "English (US)", "LTR"),
		domain.NewLanguage("zh-CN", "简体中文", "LTR"),
	}
	for _, l := range defaults {
		if err := s.repo.SaveLanguage(ctx, l); err != nil {
			return err
		}
	}
	return nil
}
