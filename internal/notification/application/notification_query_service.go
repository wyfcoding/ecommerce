package application

import (
	"context"
	"log/slog"

	"github.com/wyfcoding/ecommerce/internal/notification/domain"
)

// NotificationQueryService 处理通知的读操作。
type NotificationQueryService struct {
	repo              domain.NotificationRepository
	readRepo          domain.NotificationReadRepository
	templateReadRepo  domain.NotificationTemplateReadRepository
	searchRepo        domain.NotificationSearchRepository
	templateSearchRepo domain.NotificationTemplateSearchRepository
	logger            *slog.Logger
}

// NewNotificationQueryService 负责处理通知相关的读操作和查询逻辑。
func NewNotificationQueryService(
	repo domain.NotificationRepository,
	readRepo domain.NotificationReadRepository,
	templateReadRepo domain.NotificationTemplateReadRepository,
	searchRepo domain.NotificationSearchRepository,
	templateSearchRepo domain.NotificationTemplateSearchRepository,
	logger *slog.Logger,
) *NotificationQueryService {
	return &NotificationQueryService{
		repo:              repo,
		readRepo:          readRepo,
		templateReadRepo:  templateReadRepo,
		searchRepo:        searchRepo,
		templateSearchRepo: templateSearchRepo,
		logger:            logger,
	}
}

// GetNotification 获取指定ID的通知详情。
func (q *NotificationQueryService) GetNotification(ctx context.Context, id uint64) (*domain.Notification, error) {
	if q.readRepo != nil {
		if notif, err := q.readRepo.GetByID(ctx, id); err == nil && notif != nil {
			return notif, nil
		}
	}
	notif, err := q.repo.GetNotification(ctx, id)
	if err != nil {
		return nil, err
	}
	if notif != nil && q.readRepo != nil {
		if err := q.readRepo.Save(ctx, notif); err != nil {
			q.logger.WarnContext(ctx, "failed to warm notification cache", "id", id, "error", err)
		}
	}
	return notif, nil
}

// ListNotifications 获取用户通知列表。
func (q *NotificationQueryService) ListNotifications(ctx context.Context, userID uint64, status *int, page, pageSize int) ([]*domain.Notification, int64, error) {
	offset := (page - 1) * pageSize
	var notifStatus *domain.NotificationStatus
	if status != nil {
		s := domain.NotificationStatus(*status)
		notifStatus = &s
	}
	if q.searchRepo != nil {
		list, total, err := q.searchRepo.Search(ctx, userID, notifStatus, offset, pageSize)
		if err == nil {
			return list, total, nil
		}
		q.logger.WarnContext(ctx, "notification search fallback to mysql", "user_id", userID, "error", err)
	}
	return q.repo.ListNotifications(ctx, userID, notifStatus, offset, pageSize)
}

// GetUnreadCount 获取指定用户的未读通知数量。
func (q *NotificationQueryService) GetUnreadCount(ctx context.Context, userID uint64) (int64, error) {
	return q.repo.CountUnreadNotifications(ctx, userID)
}

// ListTemplates 获取通知模板列表（分页）。
func (q *NotificationQueryService) ListTemplates(ctx context.Context, page, pageSize int) ([]*domain.NotificationTemplate, int64, error) {
	offset := (page - 1) * pageSize
	if q.templateSearchRepo != nil {
		list, total, err := q.templateSearchRepo.Search(ctx, offset, pageSize)
		if err == nil {
			return list, total, nil
		}
		q.logger.WarnContext(ctx, "template search fallback to mysql", "error", err)
	}
	return q.repo.ListTemplates(ctx, offset, pageSize)
}
