package application

import (
	"context"

	"github.com/wyfcoding/ecommerce/internal/content_moderation/domain"
)

// ModerationService 作为内容审核操作的门面。
type ModerationService struct {
	manager *ModerationManager
	query   *ModerationQuery
}

// NewModerationService creates a new ModerationService facade.
func NewModerationService(manager *ModerationManager, query *ModerationQuery) *ModerationService {
	return &ModerationService{
		manager: manager,
		query:   query,
	}
}

// --- 写操作（委托给 Manager）---

func (s *ModerationService) SubmitContent(ctx context.Context, contentType domain.ContentType, contentID uint64, content string, userID uint64) (*domain.ModerationRecord, error) {
	return s.manager.SubmitContent(ctx, contentType, contentID, content, userID)
}

func (s *ModerationService) ReviewContent(ctx context.Context, id uint64, moderatorID uint64, approved bool, reason string) error {
	return s.manager.ReviewContent(ctx, id, moderatorID, approved, reason)
}

func (s *ModerationService) AddSensitiveWord(ctx context.Context, word, category string, level int8) (*domain.SensitiveWord, error) {
	return s.manager.AddSensitiveWord(ctx, word, category, level)
}

func (s *ModerationService) DeleteSensitiveWord(ctx context.Context, id uint64) error {
	return s.manager.DeleteSensitiveWord(ctx, id)
}

// --- 读操作（委托给 Query）---

func (s *ModerationService) ListPendingRecords(ctx context.Context, page, pageSize int) ([]*domain.ModerationRecord, int64, error) {
	return s.query.ListPendingRecords(ctx, page, pageSize)
}

func (s *ModerationService) ListSensitiveWords(ctx context.Context, page, pageSize int) ([]*domain.SensitiveWord, int64, error) {
	return s.query.ListSensitiveWords(ctx, page, pageSize)
}
