package domain

import (
	"context"
)

type I18nRepository interface {
	SaveLanguage(ctx context.Context, lang *Language) error
	ListLanguages(ctx context.Context) ([]Language, error)

	SaveTranslation(ctx context.Context, trans *Translation) error
	GetTranslation(ctx context.Context, lang, key string) (*Translation, error)
	ListTranslations(ctx context.Context, lang string, keys []string, ns string) ([]Translation, error)
}
