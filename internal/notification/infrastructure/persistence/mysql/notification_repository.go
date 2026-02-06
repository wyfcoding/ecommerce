package mysql

import (
	"context"
	"errors"

	"github.com/wyfcoding/ecommerce/internal/notification/domain"
	"gorm.io/gorm"
)

type notificationRepository struct {
	db *gorm.DB
}

// NewNotificationRepository 创建并返回一个新的 NotificationRepository 实例。
func NewNotificationRepository(db *gorm.DB) domain.NotificationRepository {
	return &notificationRepository{db: db}
}

func (r *notificationRepository) BeginTx(ctx context.Context) any {
	return r.db.WithContext(ctx).Begin()
}

func (r *notificationRepository) CommitTx(tx any) error {
	gormTx, ok := tx.(*gorm.DB)
	if !ok || gormTx == nil {
		return errors.New("invalid transaction")
	}
	return gormTx.Commit().Error
}

func (r *notificationRepository) RollbackTx(tx any) error {
	gormTx, ok := tx.(*gorm.DB)
	if !ok || gormTx == nil {
		return errors.New("invalid transaction")
	}
	return gormTx.Rollback().Error
}

func (r *notificationRepository) WithTx(ctx context.Context, fn func(tx any) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(tx)
	})
}

func (r *notificationRepository) SaveNotification(ctx context.Context, notification *domain.Notification) error {
	return r.saveNotificationWithTx(ctx, r.db, notification)
}

func (r *notificationRepository) SaveNotificationInTx(ctx context.Context, tx any, notification *domain.Notification) error {
	gormTx, ok := tx.(*gorm.DB)
	if !ok || gormTx == nil {
		return errors.New("invalid transaction")
	}
	return r.saveNotificationWithTx(ctx, gormTx, notification)
}

func (r *notificationRepository) GetNotification(ctx context.Context, id uint64) (*domain.Notification, error) {
	var model NotificationModel
	if err := r.db.WithContext(ctx).First(&model, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toNotification(&model), nil
}

func (r *notificationRepository) ListNotifications(ctx context.Context, userID uint64, status *domain.NotificationStatus, offset, limit int) ([]*domain.Notification, int64, error) {
	var list []*NotificationModel
	var total int64

	db := r.db.WithContext(ctx).Model(&NotificationModel{}).Where("user_id = ?", userID)
	if status != nil {
		db = db.Where("status = ?", *status)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := db.Offset(offset).Limit(limit).Order("created_at desc").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	items := make([]*domain.Notification, len(list))
	for i, model := range list {
		items[i] = toNotification(model)
	}

	return items, total, nil
}

func (r *notificationRepository) CountUnreadNotifications(ctx context.Context, userID uint64) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&NotificationModel{}).
		Where("user_id = ? AND status = ?", userID, domain.NotificationStatusUnread).
		Count(&count).Error
	return count, err
}

func (r *notificationRepository) DeleteNotification(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&NotificationModel{}, id).Error
}

func (r *notificationRepository) DeleteNotificationInTx(ctx context.Context, tx any, id uint64) error {
	gormTx, ok := tx.(*gorm.DB)
	if !ok || gormTx == nil {
		return errors.New("invalid transaction")
	}
	return gormTx.WithContext(ctx).Delete(&NotificationModel{}, id).Error
}

func (r *notificationRepository) SaveTemplate(ctx context.Context, template *domain.NotificationTemplate) error {
	return r.saveTemplateWithTx(ctx, r.db, template)
}

func (r *notificationRepository) SaveTemplateInTx(ctx context.Context, tx any, template *domain.NotificationTemplate) error {
	gormTx, ok := tx.(*gorm.DB)
	if !ok || gormTx == nil {
		return errors.New("invalid transaction")
	}
	return r.saveTemplateWithTx(ctx, gormTx, template)
}

func (r *notificationRepository) GetTemplateByCode(ctx context.Context, code string) (*domain.NotificationTemplate, error) {
	var model NotificationTemplateModel
	if err := r.db.WithContext(ctx).Where("code = ?", code).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toTemplate(&model), nil
}

func (r *notificationRepository) ListTemplates(ctx context.Context, offset, limit int) ([]*domain.NotificationTemplate, int64, error) {
	var list []*NotificationTemplateModel
	var total int64

	db := r.db.WithContext(ctx).Model(&NotificationTemplateModel{})
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := db.Offset(offset).Limit(limit).Order("created_at desc").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	items := make([]*domain.NotificationTemplate, len(list))
	for i, model := range list {
		items[i] = toTemplate(model)
	}

	return items, total, nil
}

func (r *notificationRepository) saveNotificationWithTx(ctx context.Context, tx *gorm.DB, notification *domain.Notification) error {
	if notification == nil {
		return nil
	}
	model := toNotificationModel(notification)
	if err := tx.WithContext(ctx).Save(model).Error; err != nil {
		return err
	}
	notification.ID = uint64(model.ID)
	notification.CreatedAt = model.CreatedAt
	notification.UpdatedAt = model.UpdatedAt
	return nil
}

func (r *notificationRepository) saveTemplateWithTx(ctx context.Context, tx *gorm.DB, template *domain.NotificationTemplate) error {
	if template == nil {
		return nil
	}
	model := toTemplateModel(template)
	if err := tx.WithContext(ctx).Save(model).Error; err != nil {
		return err
	}
	template.ID = uint64(model.ID)
	template.CreatedAt = model.CreatedAt
	template.UpdatedAt = model.UpdatedAt
	return nil
}
