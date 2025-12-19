package persistence

import (
	"context"
	"errors"

	"github.com/wyfcoding/ecommerce/internal/notification/domain" // 导入通知模块的领域层。

	"gorm.io/gorm" // 导入GORM ORM框架。
)

type notificationRepository struct {
	db *gorm.DB // GORM数据库连接实例。
}

// NewNotificationRepository 创建并返回一个新的 notificationRepository 实例。
// db: GORM数据库连接实例。
func NewNotificationRepository(db *gorm.DB) domain.NotificationRepository {
	return &notificationRepository{db: db}
}

// --- 通知 (Notification methods) ---

// SaveNotification 将通知实体保存到数据库。
// 如果通知已存在（通过ID），则更新其信息；如果不存在，则创建。
func (r *notificationRepository) SaveNotification(ctx context.Context, notification *domain.Notification) error {
	return r.db.WithContext(ctx).Save(notification).Error
}

// GetNotification 根据ID从数据库获取通知记录。
// 如果记录未找到，则返回nil。
func (r *notificationRepository) GetNotification(ctx context.Context, id uint64) (*domain.Notification, error) {
	var notification domain.Notification
	if err := r.db.WithContext(ctx).First(&notification, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // 如果记录未找到，返回nil。
		}
		return nil, err // 其他错误则返回。
	}
	return &notification, nil
}

// ListNotifications 从数据库列出指定用户ID的所有通知记录，支持通过状态过滤和分页。
func (r *notificationRepository) ListNotifications(ctx context.Context, userID uint64, status *domain.NotificationStatus, offset, limit int) ([]*domain.Notification, int64, error) {
	var list []*domain.Notification
	var total int64

	db := r.db.WithContext(ctx).Model(&domain.Notification{}).Where("user_id = ?", userID)
	if status != nil { // 如果提供了状态，则按状态过滤。
		db = db.Where("status = ?", *status)
	}

	// 统计总记录数。
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 应用分页和排序。
	if err := db.Offset(offset).Limit(limit).Order("created_at desc").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// CountUnreadNotifications 统计指定用户ID的未读通知数量。
func (r *notificationRepository) CountUnreadNotifications(ctx context.Context, userID uint64) (int64, error) {
	var count int64
	// 查询用户ID匹配且状态为NotificationStatusUnread的通知数量。
	err := r.db.WithContext(ctx).Model(&domain.Notification{}).
		Where("user_id = ? AND status = ?", userID, domain.NotificationStatusUnread).
		Count(&count).Error
	return count, err
}

// --- 模板 (NotificationTemplate methods) ---

// SaveTemplate 将通知模板实体保存到数据库。
func (r *notificationRepository) SaveTemplate(ctx context.Context, template *domain.NotificationTemplate) error {
	return r.db.WithContext(ctx).Save(template).Error
}

// GetTemplateByCode 根据模板代码从数据库获取通知模板记录。
// 如果记录未找到，则返回nil。
func (r *notificationRepository) GetTemplateByCode(ctx context.Context, code string) (*domain.NotificationTemplate, error) {
	var template domain.NotificationTemplate
	if err := r.db.WithContext(ctx).Where("code = ?", code).First(&template).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // 如果记录未找到，返回nil。
		}
		return nil, err // 其他错误则返回。
	}
	return &template, nil
}

// ListTemplates 从数据库列出所有通知模板记录，支持分页。
func (r *notificationRepository) ListTemplates(ctx context.Context, offset, limit int) ([]*domain.NotificationTemplate, int64, error) {
	var list []*domain.NotificationTemplate
	var total int64

	db := r.db.WithContext(ctx).Model(&domain.NotificationTemplate{})

	// 统计总记录数。
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 应用分页和排序。
	if err := db.Offset(offset).Limit(limit).Order("created_at desc").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}
