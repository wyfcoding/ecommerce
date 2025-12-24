package application

import (
	"context"
	"log/slog"

	"github.com/wyfcoding/ecommerce/internal/contentmoderation/domain"
)

// ModerationManager 处理内容审核的写操作。
type ModerationManager struct {
	repo   domain.ModerationRepository
	logger *slog.Logger
}

// NewModerationManager 创建并返回一个新的 ModerationManager 实例。
func NewModerationManager(repo domain.ModerationRepository, logger *slog.Logger) *ModerationManager {
	return &ModerationManager{
		repo:   repo,
		logger: logger,
	}
}

// SubmitContent 提交内容进行审核。
func (m *ModerationManager) SubmitContent(ctx context.Context, contentType domain.ContentType, contentID uint64, content string, userID uint64) (*domain.ModerationRecord, error) {
	record := domain.NewModerationRecord(contentType, contentID, content, userID)
	record.SetAIResult(0.1, []string{"safe"})

	if err := m.repo.CreateRecord(ctx, record); err != nil {
		m.logger.ErrorContext(ctx, "failed to create moderation record", "content_type", contentType, "content_id", contentID, "error", err)
		return nil, err
	}
	m.logger.InfoContext(ctx, "moderation record created successfully", "record_id", record.ID, "content_type", contentType, "content_id", contentID)
	return record, nil
}

// ReviewContent 对内容进行人工审核。
func (m *ModerationManager) ReviewContent(ctx context.Context, id uint64, moderatorID uint64, approved bool, reason string) error {
	record, err := m.repo.GetRecord(ctx, id)
	if err != nil {
		return err
	}

	if approved {
		record.Approve(moderatorID)
	} else {
		record.Reject(moderatorID, reason)
	}

	return m.repo.UpdateRecord(ctx, record)
}

// AddSensitiveWord 添加一个敏感词到系统。
func (m *ModerationManager) AddSensitiveWord(ctx context.Context, word, category string, level int8) (*domain.SensitiveWord, error) {
	sensitiveWord := domain.NewSensitiveWord(word, category, level)
	if err := m.repo.CreateWord(ctx, sensitiveWord); err != nil {
		m.logger.ErrorContext(ctx, "failed to create sensitive word", "word", word, "error", err)
		return nil, err
	}
	m.logger.InfoContext(ctx, "sensitive word created successfully", "word_id", sensitiveWord.ID, "word", word)
	return sensitiveWord, nil
}

// DeleteSensitiveWord 根据ID删除一个敏感词。
func (m *ModerationManager) DeleteSensitiveWord(ctx context.Context, id uint64) error {
	return m.repo.DeleteWord(ctx, id)
}
