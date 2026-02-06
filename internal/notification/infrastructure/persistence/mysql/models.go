package mysql

import (
	"time"

	"github.com/wyfcoding/ecommerce/internal/notification/domain"
	"gorm.io/gorm"
)

// NotificationModel 通知写模型。
type NotificationModel struct {
	gorm.Model
	UserID    uint64                     `gorm:"not null;index;comment:用户ID"`
	NotifType domain.NotificationType    `gorm:"type:varchar(32);not null;comment:通知类型"`
	Channel   domain.NotificationChannel `gorm:"type:varchar(32);not null;comment:通知渠道"`
	Title     string                     `gorm:"type:varchar(255);not null;comment:标题"`
	Content   string                     `gorm:"type:text;not null;comment:内容"`
	Data      domain.JSONMap             `gorm:"type:json;comment:附加数据"`
	Status    domain.NotificationStatus  `gorm:"not null;default:0;comment:状态"`
	ReadAt    *time.Time                 `gorm:"comment:阅读时间"`
}

// NotificationTemplateModel 通知模板写模型。
type NotificationTemplateModel struct {
	gorm.Model
	Code      string                     `gorm:"type:varchar(64);uniqueIndex;not null;comment:模板编码"`
	Name      string                     `gorm:"type:varchar(128);comment:模板名称"`
	NotifType domain.NotificationType    `gorm:"type:varchar(32);not null;comment:通知类型"`
	Channel   domain.NotificationChannel `gorm:"type:varchar(32);not null;comment:通知渠道"`
	Title     string                     `gorm:"type:varchar(255);not null;comment:标题"`
	Content   string                     `gorm:"type:text;not null;comment:内容"`
	Variables domain.StringArray         `gorm:"type:json;comment:变量列表"`
	Enabled   bool                       `gorm:"default:true;comment:是否启用"`
}

func (NotificationModel) TableName() string {
	return "notifications"
}

func (NotificationTemplateModel) TableName() string {
	return "notification_templates"
}

func toNotificationModel(n *domain.Notification) *NotificationModel {
	if n == nil {
		return nil
	}
	return &NotificationModel{
		Model: gorm.Model{
			ID:        uint(n.ID),
			CreatedAt: n.CreatedAt,
			UpdatedAt: n.UpdatedAt,
		},
		UserID:    n.UserID,
		NotifType: n.NotifType,
		Channel:   n.Channel,
		Title:     n.Title,
		Content:   n.Content,
		Data:      n.Data,
		Status:    n.Status,
		ReadAt:    n.ReadAt,
	}
}

func toNotification(model *NotificationModel) *domain.Notification {
	if model == nil {
		return nil
	}
	return &domain.Notification{
		ID:        uint64(model.ID),
		CreatedAt: model.CreatedAt,
		UpdatedAt: model.UpdatedAt,
		UserID:    model.UserID,
		NotifType: model.NotifType,
		Channel:   model.Channel,
		Title:     model.Title,
		Content:   model.Content,
		Data:      model.Data,
		Status:    model.Status,
		ReadAt:    model.ReadAt,
	}
}

func toTemplateModel(t *domain.NotificationTemplate) *NotificationTemplateModel {
	if t == nil {
		return nil
	}
	return &NotificationTemplateModel{
		Model: gorm.Model{
			ID:        uint(t.ID),
			CreatedAt: t.CreatedAt,
			UpdatedAt: t.UpdatedAt,
		},
		Code:      t.Code,
		Name:      t.Name,
		NotifType: t.NotifType,
		Channel:   t.Channel,
		Title:     t.Title,
		Content:   t.Content,
		Variables: t.Variables,
		Enabled:   t.Enabled,
	}
}

func toTemplate(model *NotificationTemplateModel) *domain.NotificationTemplate {
	if model == nil {
		return nil
	}
	return &domain.NotificationTemplate{
		ID:        uint64(model.ID),
		CreatedAt: model.CreatedAt,
		UpdatedAt: model.UpdatedAt,
		Code:      model.Code,
		Name:      model.Name,
		NotifType: model.NotifType,
		Channel:   model.Channel,
		Title:     model.Title,
		Content:   model.Content,
		Variables: model.Variables,
		Enabled:   model.Enabled,
	}
}
