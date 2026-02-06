// 生成摘要：新增通知读模型投影服务，消费事件后刷新 Redis/ES 读侧。
// 假设：读模型以 notification_id 与 template_code 为主键，写模型为最终一致性来源。
package application

import (
	"context"
	"log/slog"

	"github.com/wyfcoding/ecommerce/internal/notification/domain"
)

// NotificationProjectionService 负责将事件转换为读模型更新。
type NotificationProjectionService struct {
	repo               domain.NotificationRepository
	readRepo           domain.NotificationReadRepository
	templateReadRepo   domain.NotificationTemplateReadRepository
	searchRepo         domain.NotificationSearchRepository
	templateSearchRepo domain.NotificationTemplateSearchRepository
	logger             *slog.Logger
}

// NewNotificationProjectionService 创建通知投影服务。
func NewNotificationProjectionService(
	repo domain.NotificationRepository,
	readRepo domain.NotificationReadRepository,
	templateReadRepo domain.NotificationTemplateReadRepository,
	searchRepo domain.NotificationSearchRepository,
	templateSearchRepo domain.NotificationTemplateSearchRepository,
	logger *slog.Logger,
) *NotificationProjectionService {
	return &NotificationProjectionService{
		repo:               repo,
		readRepo:           readRepo,
		templateReadRepo:   templateReadRepo,
		searchRepo:         searchRepo,
		templateSearchRepo: templateSearchRepo,
		logger:             logger,
	}
}

// OnNotificationCreated 处理通知创建事件。
func (s *NotificationProjectionService) OnNotificationCreated(ctx context.Context, event *domain.NotificationCreatedEvent) error {
	if event == nil {
		return nil
	}
	return s.refreshNotification(ctx, event.NotificationID)
}

// OnNotificationRead 处理通知已读事件。
func (s *NotificationProjectionService) OnNotificationRead(ctx context.Context, event *domain.NotificationReadEvent) error {
	if event == nil {
		return nil
	}
	return s.refreshNotification(ctx, event.NotificationID)
}

// OnNotificationDeleted 处理通知删除事件。
func (s *NotificationProjectionService) OnNotificationDeleted(ctx context.Context, event *domain.NotificationDeletedEvent) error {
	if event == nil {
		return nil
	}
	if s.readRepo != nil {
		_ = s.readRepo.Delete(ctx, event.NotificationID)
	}
	if s.searchRepo != nil {
		_ = s.searchRepo.Delete(ctx, event.NotificationID)
	}
	return nil
}

// OnTemplateCreated 处理模板创建事件。
func (s *NotificationProjectionService) OnTemplateCreated(ctx context.Context, event *domain.NotificationTemplateCreatedEvent) error {
	if event == nil {
		return nil
	}
	template, err := s.repo.GetTemplateByCode(ctx, event.Code)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to load template for projection", "code", event.Code, "error", err)
		return err
	}
	if template == nil {
		if s.templateReadRepo != nil {
			_ = s.templateReadRepo.DeleteByCode(ctx, event.Code)
		}
		if s.templateSearchRepo != nil {
			_ = s.templateSearchRepo.Delete(ctx, event.TemplateID)
		}
		return nil
	}
	if s.templateReadRepo != nil {
		if err := s.templateReadRepo.Save(ctx, template); err != nil {
			s.logger.ErrorContext(ctx, "failed to save template read model", "code", event.Code, "error", err)
			return err
		}
	}
	if s.templateSearchRepo != nil {
		if err := s.templateSearchRepo.Index(ctx, template); err != nil {
			s.logger.ErrorContext(ctx, "failed to index template", "code", event.Code, "error", err)
			return err
		}
	}
	return nil
}

func (s *NotificationProjectionService) refreshNotification(ctx context.Context, id uint64) error {
	if id == 0 {
		return nil
	}
	notif, err := s.repo.GetNotification(ctx, id)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to load notification for projection", "id", id, "error", err)
		return err
	}
	if notif == nil {
		if s.readRepo != nil {
			_ = s.readRepo.Delete(ctx, id)
		}
		if s.searchRepo != nil {
			_ = s.searchRepo.Delete(ctx, id)
		}
		return nil
	}
	if s.readRepo != nil {
		if err := s.readRepo.Save(ctx, notif); err != nil {
			s.logger.ErrorContext(ctx, "failed to save notification read model", "id", id, "error", err)
			return err
		}
	}
	if s.searchRepo != nil {
		if err := s.searchRepo.Index(ctx, notif); err != nil {
			s.logger.ErrorContext(ctx, "failed to index notification", "id", id, "error", err)
			return err
		}
	}
	return nil
}
