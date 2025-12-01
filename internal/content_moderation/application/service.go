package application

import (
	"context"

	"github.com/wyfcoding/ecommerce/internal/content_moderation/domain/entity"
	"github.com/wyfcoding/ecommerce/internal/content_moderation/domain/repository"

	"log/slog"
)

type ModerationService struct {
	repo   repository.ModerationRepository
	logger *slog.Logger
}

func NewModerationService(repo repository.ModerationRepository, logger *slog.Logger) *ModerationService {
	return &ModerationService{
		repo:   repo,
		logger: logger,
	}
}

// SubmitContent 提交内容审核
func (s *ModerationService) SubmitContent(ctx context.Context, contentType entity.ContentType, contentID uint64, content string, userID uint64) (*entity.ModerationRecord, error) {
	record := entity.NewModerationRecord(contentType, contentID, content, userID)

	// TODO: Call AI service for pre-moderation
	// For now, mock AI result
	record.SetAIResult(0.1, []string{"safe"})

	if err := s.repo.CreateRecord(ctx, record); err != nil {
		s.logger.ErrorContext(ctx, "failed to create moderation record", "content_type", contentType, "content_id", contentID, "error", err)
		return nil, err
	}
	s.logger.InfoContext(ctx, "moderation record created successfully", "record_id", record.ID, "content_type", contentType, "content_id", contentID)
	return record, nil
}

// ReviewContent 人工审核
func (s *ModerationService) ReviewContent(ctx context.Context, id uint64, moderatorID uint64, approved bool, reason string) error {
	record, err := s.repo.GetRecord(ctx, id)
	if err != nil {
		return err
	}

	if approved {
		record.Approve(moderatorID)
	} else {
		record.Reject(moderatorID, reason)
	}

	return s.repo.UpdateRecord(ctx, record)
}

// ListPendingRecords 获取待审核列表
func (s *ModerationService) ListPendingRecords(ctx context.Context, page, pageSize int) ([]*entity.ModerationRecord, int64, error) {
	offset := (page - 1) * pageSize
	return s.repo.ListRecords(ctx, entity.ModerationStatusPending, offset, pageSize)
}

// AddSensitiveWord 添加敏感词
func (s *ModerationService) AddSensitiveWord(ctx context.Context, word, category string, level int8) (*entity.SensitiveWord, error) {
	sensitiveWord := entity.NewSensitiveWord(word, category, level)
	if err := s.repo.CreateWord(ctx, sensitiveWord); err != nil {
		s.logger.ErrorContext(ctx, "failed to create sensitive word", "word", word, "error", err)
		return nil, err
	}
	s.logger.InfoContext(ctx, "sensitive word created successfully", "word_id", sensitiveWord.ID, "word", word)
	return sensitiveWord, nil
}

// ListSensitiveWords 获取敏感词列表
func (s *ModerationService) ListSensitiveWords(ctx context.Context, page, pageSize int) ([]*entity.SensitiveWord, int64, error) {
	offset := (page - 1) * pageSize
	return s.repo.ListWords(ctx, offset, pageSize)
}

// DeleteSensitiveWord 删除敏感词
func (s *ModerationService) DeleteSensitiveWord(ctx context.Context, id uint64) error {
	return s.repo.DeleteWord(ctx, id)
}
