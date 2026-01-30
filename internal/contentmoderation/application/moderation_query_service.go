package application

import (
	"context"

	"github.com/wyfcoding/ecommerce/internal/contentmoderation/domain"
)

// ModerationQueryService 处理内容审核的读操作。
type ModerationQueryService struct {
	repo domain.ModerationRepository
}

// NewModerationQueryService 创建并返回一个新的 ModerationQueryService 实例。
func NewModerationQueryService(repo domain.ModerationRepository) *ModerationQueryService {
	return &ModerationQueryService{
		repo: repo,
	}
}

// ListPendingRecords 获取所有待人工审核的内容记录列表。
func (q *ModerationQueryService) ListPendingRecords(ctx context.Context, page, pageSize int) ([]*domain.ModerationRecord, int64, error) {
	offset := (page - 1) * pageSize
	return q.repo.ListRecords(ctx, domain.ModerationStatusPending, offset, pageSize)
}

// ListSensitiveWords 获取敏感词列表。
func (q *ModerationQueryService) ListSensitiveWords(ctx context.Context, page, pageSize int) ([]*domain.SensitiveWord, int64, error) {
	offset := (page - 1) * pageSize
	return q.repo.ListWords(ctx, offset, pageSize)
}
