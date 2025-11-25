package repository

import (
	"context"
	"github.com/wyfcoding/ecommerce/internal/content_moderation/domain/entity"
)

// ModerationRepository 内容审核仓储接口
type ModerationRepository interface {
	// ModerationRecord methods
	CreateRecord(ctx context.Context, record *entity.ModerationRecord) error
	GetRecord(ctx context.Context, id uint64) (*entity.ModerationRecord, error)
	UpdateRecord(ctx context.Context, record *entity.ModerationRecord) error
	ListRecords(ctx context.Context, status entity.ModerationStatus, offset, limit int) ([]*entity.ModerationRecord, int64, error)

	// SensitiveWord methods
	CreateWord(ctx context.Context, word *entity.SensitiveWord) error
	GetWord(ctx context.Context, id uint64) (*entity.SensitiveWord, error)
	UpdateWord(ctx context.Context, word *entity.SensitiveWord) error
	DeleteWord(ctx context.Context, id uint64) error
	ListWords(ctx context.Context, offset, limit int) ([]*entity.SensitiveWord, int64, error)
	FindWord(ctx context.Context, word string) (*entity.SensitiveWord, error)
}
