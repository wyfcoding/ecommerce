package domain

import (
	"context"
)

// ModerationRepository 是内容审核模块的仓储接口。
type ModerationRepository interface {
	// --- ModerationRecord methods ---
	CreateRecord(ctx context.Context, record *ModerationRecord) error
	GetRecord(ctx context.Context, id uint64) (*ModerationRecord, error)
	UpdateRecord(ctx context.Context, record *ModerationRecord) error
	ListRecords(ctx context.Context, status ModerationStatus, offset, limit int) ([]*ModerationRecord, int64, error)

	// --- SensitiveWord methods ---
	CreateWord(ctx context.Context, word *SensitiveWord) error
	GetWord(ctx context.Context, id uint64) (*SensitiveWord, error)
	UpdateWord(ctx context.Context, word *SensitiveWord) error
	DeleteWord(ctx context.Context, id uint64) error
	ListWords(ctx context.Context, offset, limit int) ([]*SensitiveWord, int64, error)
	FindWord(ctx context.Context, word string) (*SensitiveWord, error)
}
