package data

import (
	"context"
	"ecommerce/internal/notification/biz"
	"ecommerce/internal/notification/data/model"

	"gorm.io/gorm"
)

type notificationRepo struct {
	data *Data
}

// NewNotificationRepo creates a new NotificationRepo.
func NewNotificationRepo(data *Data) biz.NotificationRepo {
	return &notificationRepo{data: data}
}

// CreateNotification creates a new notification record.
func (r *notificationRepo) CreateNotification(ctx context.Context, notification *biz.Notification) (*biz.Notification, error) {
	po := &model.Notification{
		NotificationID: notification.NotificationID,
		UserID:         notification.UserID,
		Type:           notification.Type,
		Title:          notification.Title,
		Content:        notification.Content,
		IsRead:         notification.IsRead,
	}
	if err := r.data.db.WithContext(ctx).Create(po).Error; err != nil {
		return nil, err
	}
	notification.ID = po.ID
	return notification, nil
}

// GetNotificationByID retrieves a notification by its ID.
func (r *notificationRepo) GetNotificationByID(ctx context.Context, notificationID string) (*biz.Notification, error) {
	var po model.Notification
	if err := r.data.db.WithContext(ctx).Where("notification_id = ?", notificationID).First(&po).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // Notification not found
		}
		return nil, err
	}
	return &biz.Notification{
		ID:             po.ID,
		NotificationID: po.NotificationID,
		UserID:         po.UserID,
		Type:           po.Type,
		Title:          po.Title,
		Content:        po.Content,
		IsRead:         po.IsRead,
		CreatedAt:      po.CreatedAt,
	}, nil
}

// ListNotificationsByUserID lists notifications for a specific user.
func (r *notificationRepo) ListNotificationsByUserID(ctx context.Context, userID uint64, includeRead bool, pageSize, pageNum uint32) ([]*biz.Notification, uint64, error) {
	var notifications []*model.Notification
	var total int64
	query := r.data.db.WithContext(ctx).Where("user_id = ?", userID)

	if !includeRead {
		query = query.Where("is_read = ?", false)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (pageNum - 1) * pageSize
	if err := query.Offset(int(offset)).Limit(int(pageSize)).Order("created_at DESC").Find(&notifications).Error; err != nil {
		return nil, 0, err
	}

	bizNotifications := make([]*biz.Notification, len(notifications))
	for i, notif := range notifications {
		bizNotifications[i] = &biz.Notification{
			ID:             notif.ID,
			NotificationID: notif.NotificationID,
			UserID:         notif.UserID,
			Type:           notif.Type,
			Title:          notif.Title,
			Content:        notif.Content,
			IsRead:         notif.IsRead,
			CreatedAt:      notif.CreatedAt,
		}
	}
	return bizNotifications, uint64(total), nil
}

// MarkNotificationAsRead marks a notification as read.
func (r *notificationRepo) MarkNotificationAsRead(ctx context.Context, notificationID string) error {
	return r.data.db.WithContext(ctx).Model(&model.Notification{}).Where("notification_id = ?", notificationID).Update("is_read", true).Error
}
