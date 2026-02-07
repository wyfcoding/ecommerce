package domain

import "context"

// ModerationRecordSearchRepository 定义审核记录搜索仓储接口（Elasticsearch）。
type ModerationRecordSearchRepository interface {
	Index(ctx context.Context, record *ModerationRecord) error
	Delete(ctx context.Context, recordID uint64) error
	Search(ctx context.Context, query *ModerationRecordQuery, offset, limit int) ([]*ModerationRecord, int64, error)
}
