package application

import (
	"context"

	"github.com/wyfcoding/ecommerce/internal/content_moderation/domain"
)

// ModerationQuery 处理内容审核的读操作。
type ModerationQuery struct {
	repo domain.ModerationRepository
}

// NewModerationQuery creates a new ModerationQuery instance.
func NewModerationQuery(repo domain.ModerationRepository) *ModerationQuery {
	return &ModerationQuery{
		repo: repo,
	}
}

// ListPendingRecords 获取所有待人工审核的内容记录列表。
func (q *ModerationQuery) ListPendingRecords(ctx context.Context, page, pageSize int) ([]*domain.ModerationRecord, int64, error) {
	offset := (page - 1) * pageSize
	return q.repo.ListRecords(ctx, domain.ModerationStatusPending, offset, pageSize)
}

// ListSensitiveWords 获取敏感词列表。
func (q *ModerationQuery) ListSensitiveWords(ctx context.Context, page, pageSize int) ([]*domain.SensitiveWord, int64, error) {
	offset := (page - 1) * pageSize
	return q.repo.ListWords(ctx, offset, pageSize)
}
