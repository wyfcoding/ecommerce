// 生成摘要：新增内容审核读模型投影服务，消费事件后刷新 Redis/ES 读侧。
package application

import (
	"context"
	"log/slog"

	"github.com/wyfcoding/ecommerce/internal/contentmoderation/domain"
)

// ModerationProjectionService 负责将内容审核事件投影到读模型。
type ModerationProjectionService struct {
	repo             domain.ModerationRepository
	recordReadRepo   domain.ModerationRecordReadRepository
	wordReadRepo     domain.SensitiveWordReadRepository
	recordSearchRepo domain.ModerationRecordSearchRepository
	logger           *slog.Logger
}

// NewModerationProjectionService 创建投影服务。
func NewModerationProjectionService(
	repo domain.ModerationRepository,
	recordReadRepo domain.ModerationRecordReadRepository,
	wordReadRepo domain.SensitiveWordReadRepository,
	recordSearchRepo domain.ModerationRecordSearchRepository,
	logger *slog.Logger,
) *ModerationProjectionService {
	return &ModerationProjectionService{
		repo:             repo,
		recordReadRepo:   recordReadRepo,
		wordReadRepo:     wordReadRepo,
		recordSearchRepo: recordSearchRepo,
		logger:           logger,
	}
}

func (s *ModerationProjectionService) OnRecordCreated(ctx context.Context, event *domain.ModerationRecordCreatedEvent) error {
	if event == nil {
		return nil
	}
	return s.refreshRecord(ctx, event.RecordID)
}

func (s *ModerationProjectionService) OnRecordUpdated(ctx context.Context, event *domain.ModerationRecordUpdatedEvent) error {
	if event == nil {
		return nil
	}
	return s.refreshRecord(ctx, event.RecordID)
}

func (s *ModerationProjectionService) OnRecordDeleted(ctx context.Context, event *domain.ModerationRecordDeletedEvent) error {
	if event == nil {
		return nil
	}
	if s.recordReadRepo != nil {
		_ = s.recordReadRepo.Delete(ctx, event.RecordID)
	}
	if s.recordSearchRepo != nil {
		_ = s.recordSearchRepo.Delete(ctx, event.RecordID)
	}
	return nil
}

func (s *ModerationProjectionService) OnSensitiveWordCreated(ctx context.Context, event *domain.SensitiveWordCreatedEvent) error {
	if event == nil {
		return nil
	}
	return s.refreshWord(ctx, event.WordID)
}

func (s *ModerationProjectionService) OnSensitiveWordUpdated(ctx context.Context, event *domain.SensitiveWordUpdatedEvent) error {
	if event == nil {
		return nil
	}
	return s.refreshWord(ctx, event.WordID)
}

func (s *ModerationProjectionService) OnSensitiveWordDeleted(ctx context.Context, event *domain.SensitiveWordDeletedEvent) error {
	if event == nil {
		return nil
	}
	if s.wordReadRepo != nil {
		_ = s.wordReadRepo.Delete(ctx, event.WordID)
	}
	return nil
}

func (s *ModerationProjectionService) refreshRecord(ctx context.Context, recordID uint64) error {
	if s.recordReadRepo == nil && s.recordSearchRepo == nil {
		return nil
	}
	record, err := s.repo.GetRecord(ctx, recordID)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to load moderation record for projection", "record_id", recordID, "error", err)
		return err
	}
	if record == nil {
		if s.recordReadRepo != nil {
			_ = s.recordReadRepo.Delete(ctx, recordID)
		}
		if s.recordSearchRepo != nil {
			_ = s.recordSearchRepo.Delete(ctx, recordID)
		}
		return nil
	}
	if s.recordReadRepo != nil {
		if err := s.recordReadRepo.Save(ctx, record); err != nil {
			s.logger.ErrorContext(ctx, "failed to save moderation record cache", "record_id", recordID, "error", err)
			return err
		}
	}
	if s.recordSearchRepo != nil {
		if err := s.recordSearchRepo.Index(ctx, record); err != nil {
			s.logger.ErrorContext(ctx, "failed to index moderation record", "record_id", recordID, "error", err)
			return err
		}
	}
	return nil
}

func (s *ModerationProjectionService) refreshWord(ctx context.Context, wordID uint64) error {
	if s.wordReadRepo == nil {
		return nil
	}
	word, err := s.repo.GetWord(ctx, wordID)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to load sensitive word for projection", "word_id", wordID, "error", err)
		return err
	}
	if word == nil {
		_ = s.wordReadRepo.Delete(ctx, wordID)
		return nil
	}
	if err := s.wordReadRepo.Save(ctx, word); err != nil {
		s.logger.ErrorContext(ctx, "failed to save sensitive word cache", "word_id", wordID, "error", err)
		return err
	}
	return nil
}
