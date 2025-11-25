package persistence

import (
	"context"
	"errors"
	"github.com/wyfcoding/ecommerce/internal/notification/domain/entity"
	"github.com/wyfcoding/ecommerce/internal/notification/domain/repository"

	"gorm.io/gorm"
)

type notificationRepository struct {
	db *gorm.DB
}

func NewNotificationRepository(db *gorm.DB) repository.NotificationRepository {
	return &notificationRepository{db: db}
}

// 通知
func (r *notificationRepository) SaveNotification(ctx context.Context, notification *entity.Notification) error {
	return r.db.WithContext(ctx).Save(notification).Error
}

func (r *notificationRepository) GetNotification(ctx context.Context, id uint64) (*entity.Notification, error) {
	var notification entity.Notification
	if err := r.db.WithContext(ctx).First(&notification, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &notification, nil
}

func (r *notificationRepository) ListNotifications(ctx context.Context, userID uint64, status *entity.NotificationStatus, offset, limit int) ([]*entity.Notification, int64, error) {
	var list []*entity.Notification
	var total int64

	db := r.db.WithContext(ctx).Model(&entity.Notification{}).Where("user_id = ?", userID)
	if status != nil {
		db = db.Where("status = ?", *status)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := db.Offset(offset).Limit(limit).Order("created_at desc").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

func (r *notificationRepository) CountUnreadNotifications(ctx context.Context, userID uint64) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&entity.Notification{}).
		Where("user_id = ? AND status = ?", userID, entity.NotificationStatusUnread).
		Count(&count).Error
	return count, err
}

// 模板
func (r *notificationRepository) SaveTemplate(ctx context.Context, template *entity.NotificationTemplate) error {
	return r.db.WithContext(ctx).Save(template).Error
}

func (r *notificationRepository) GetTemplateByCode(ctx context.Context, code string) (*entity.NotificationTemplate, error) {
	var template entity.NotificationTemplate
	if err := r.db.WithContext(ctx).Where("code = ?", code).First(&template).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &template, nil
}

func (r *notificationRepository) ListTemplates(ctx context.Context, offset, limit int) ([]*entity.NotificationTemplate, int64, error) {
	var list []*entity.NotificationTemplate
	var total int64

	db := r.db.WithContext(ctx).Model(&entity.NotificationTemplate{})
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := db.Offset(offset).Limit(limit).Order("created_at desc").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}
