package application

import (
	"context"
	"log/slog"

	"github.com/wyfcoding/ecommerce/internal/contentmoderation/domain"
)

// ModerationQueryService 处理内容审核的读操作。
type ModerationQueryService struct {
	repo             domain.ModerationRepository
	recordReadRepo   domain.ModerationRecordReadRepository
	wordReadRepo     domain.SensitiveWordReadRepository
	recordSearchRepo domain.ModerationRecordSearchRepository
	logger           *slog.Logger
}

// NewModerationQueryService 创建并返回一个新的 ModerationQueryService 实例。
func NewModerationQueryService(
	repo domain.ModerationRepository,
	recordReadRepo domain.ModerationRecordReadRepository,
	wordReadRepo domain.SensitiveWordReadRepository,
	recordSearchRepo domain.ModerationRecordSearchRepository,
	logger *slog.Logger,
) *ModerationQueryService {
	return &ModerationQueryService{
		repo:             repo,
		recordReadRepo:   recordReadRepo,
		wordReadRepo:     wordReadRepo,
		recordSearchRepo: recordSearchRepo,
		logger:           logger,
	}
}

// GetRecord 根据ID获取审核记录。
func (q *ModerationQueryService) GetRecord(ctx context.Context, id uint64) (*domain.ModerationRecord, error) {
	if q.recordReadRepo != nil {
		if cached, err := q.recordReadRepo.GetByID(ctx, id); err == nil && cached != nil {
			return cached, nil
		}
	}
	record, err := q.repo.GetRecord(ctx, id)
	if err != nil {
		return nil, err
	}
	if record != nil && q.recordReadRepo != nil {
		_ = q.recordReadRepo.Save(ctx, record)
	}
	return record, nil
}

// ListRecords 获取审核记录列表（优先 ES 搜索）。
func (q *ModerationQueryService) ListRecords(ctx context.Context, query *domain.ModerationRecordQuery) ([]*domain.ModerationRecord, int64, error) {
	page := 1
	pageSize := 20
	if query != nil {
		if query.Page > 0 {
			page = query.Page
		}
		if query.PageSize > 0 {
			pageSize = query.PageSize
		}
	}
	offset := (page - 1) * pageSize

	if q.recordSearchRepo != nil {
		list, total, err := q.recordSearchRepo.Search(ctx, query, offset, pageSize)
		if err == nil {
			return list, total, nil
		}
		q.logger.WarnContext(ctx, "moderation record search fallback to mysql", "error", err)
	}
	return q.repo.ListRecords(ctx, query)
}

// ListPendingRecords 获取所有待人工审核的内容记录列表。
func (q *ModerationQueryService) ListPendingRecords(ctx context.Context, page, pageSize int) ([]*domain.ModerationRecord, int64, error) {
	status := domain.ModerationStatusPending
	query := &domain.ModerationRecordQuery{
		Status:   &status,
		Page:     page,
		PageSize: pageSize,
	}
	return q.ListRecords(ctx, query)
}

// GetSensitiveWord 根据ID获取敏感词。
func (q *ModerationQueryService) GetSensitiveWord(ctx context.Context, id uint64) (*domain.SensitiveWord, error) {
	if q.wordReadRepo != nil {
		if cached, err := q.wordReadRepo.GetByID(ctx, id); err == nil && cached != nil {
			return cached, nil
		}
	}
	word, err := q.repo.GetWord(ctx, id)
	if err != nil {
		return nil, err
	}
	if word != nil && q.wordReadRepo != nil {
		_ = q.wordReadRepo.Save(ctx, word)
	}
	return word, nil
}

// ListSensitiveWords 获取敏感词列表。
func (q *ModerationQueryService) ListSensitiveWords(ctx context.Context, page, pageSize int) ([]*domain.SensitiveWord, int64, error) {
	offset := (page - 1) * pageSize
	return q.repo.ListWords(ctx, offset, pageSize)
}
