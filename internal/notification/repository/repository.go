package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"ecommerce/internal/notification/model"
)

// NotificationRepository 定义了通知数据仓库的接口
type NotificationRepository interface {
	CreateLog(ctx context.Context, log *model.NotificationLog) error
	UpdateLogStatus(ctx context.Context, logID uint, status model.NotificationStatus, failureReason string) error
	GetTemplate(ctx context.Context, templateID string, language string) (*model.NotificationTemplate, error)
}

// notificationRepository 是接口的具体实现
type notificationRepository struct {
	db *gorm.DB
}

// NewNotificationRepository 创建一个新的 notificationRepository 实例
func NewNotificationRepository(db *gorm.DB) NotificationRepository {
	return &notificationRepository{db: db}
}

// CreateLog 在数据库中创建一条新的通知发送日志
func (r *notificationRepository) CreateLog(ctx context.Context, log *model.NotificationLog) error {
	if err := r.db.WithContext(ctx).Create(log).Error; err != nil {
		return fmt.Errorf("数据库创建通知日志失败: %w", err)
	}
	return nil
}

// UpdateLogStatus 更新指定日志的发送状态
func (r *notificationRepository) UpdateLogStatus(ctx context.Context, logID uint, status model.NotificationStatus, failureReason string) error {
	updates := map[string]interface{}{
		"status":         status,
		"failure_reason": failureReason,
	}
	if status == model.StatusSent {
		updates["sent_at"] = gorm.Expr("NOW()")
	}

	result := r.db.WithContext(ctx).Model(&model.NotificationLog{}).Where("id = ?", logID).Updates(updates)
	if result.Error != nil {
		return fmt.Errorf("数据库更新日志状态失败: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("找不到要更新的日志记录, ID: %d", logID)
	}
	return nil
}

// GetTemplate 从数据库中获取通知模板
// 它会首先尝试获取指定语言的模板，如果找不到，则回退到默认语言 (例如 'en-US')
func (r *notificationRepository) GetTemplate(ctx context.Context, templateID string, language string) (*model.NotificationTemplate, error) {
	var template model.NotificationTemplate

	// 1. 尝试获取指定语言的模板
	err := r.db.WithContext(ctx).Where("id = ? AND language = ?", templateID, language).First(&template).Error
	if err == nil {
		return &template, nil // 找到了
	}

	// 2. 如果不是记录未找到的错误，则直接返回错误
	if err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("数据库查询模板失败: %w", err)
	}

	// 3. 如果指定语言的模板未找到，尝试获取默认语言 (e.g., en-US) 的模板
	// 这里的默认语言可以做成可配置的
	defaultLang := "en-US"
	if language != defaultLang {
		err = r.db.WithContext(ctx).Where("id = ? AND language = ?", templateID, defaultLang).First(&template).Error
		if err == nil {
			return &template, nil
		}
	}

	// 4. 如果都找不到，则返回记录未找到错误
	if err == gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("模板 '%s' 在任何语言版本中都未找到", templateID)
	}

	return nil, fmt.Errorf("数据库查询默认模板失败: %w", err)
}
