package domain

import "context"

// ModerationRecordReadRepository 定义审核记录读模型仓储接口（Redis）。
type ModerationRecordReadRepository interface {
	Save(ctx context.Context, record *ModerationRecord) error
	GetByID(ctx context.Context, id uint64) (*ModerationRecord, error)
	Delete(ctx context.Context, id uint64) error
}
