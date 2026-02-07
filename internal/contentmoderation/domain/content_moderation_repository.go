package domain

import (
	"context"
	"time"
)

// ModerationRepository 是内容审核模块的写模型仓储接口。
type ModerationRepository interface {
	// --- tx helpers ---
	BeginTx(ctx context.Context) any
	CommitTx(tx any) error
	RollbackTx(tx any) error
	WithTx(ctx context.Context, fn func(tx any) error) error

	// --- ModerationRecord methods ---
	SaveRecord(ctx context.Context, record *ModerationRecord) error
	SaveRecordInTx(ctx context.Context, tx any, record *ModerationRecord) error
	GetRecord(ctx context.Context, id uint64) (*ModerationRecord, error)
	ListRecords(ctx context.Context, query *ModerationRecordQuery) ([]*ModerationRecord, int64, error)
	DeleteRecord(ctx context.Context, id uint64) error
	DeleteRecordInTx(ctx context.Context, tx any, id uint64) error

	// --- SensitiveWord methods ---
	SaveWord(ctx context.Context, word *SensitiveWord) error
	SaveWordInTx(ctx context.Context, tx any, word *SensitiveWord) error
	GetWord(ctx context.Context, id uint64) (*SensitiveWord, error)
	ListWords(ctx context.Context, offset, limit int) ([]*SensitiveWord, int64, error)
	DeleteWord(ctx context.Context, id uint64) error
	DeleteWordInTx(ctx context.Context, tx any, id uint64) error
	FindWord(ctx context.Context, word string) (*SensitiveWord, error)
}

// ModerationRecordQuery 定义审核记录的查询条件。
type ModerationRecordQuery struct {
	UserID      uint64
	ContentType ContentType
	ContentID   uint64
	Status      *ModerationStatus
	ModeratorID uint64
	StartTime   time.Time
	EndTime     time.Time
	Page        int
	PageSize    int
}
