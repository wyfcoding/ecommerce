package domain

import "time"

const (
	NotificationCreatedEventType  = "notification.created"
	NotificationReadEventType     = "notification.read"
	NotificationDeletedEventType  = "notification.deleted"
	NotificationTemplateCreatedEventType = "notification.template.created"
)

// NotificationCreatedEvent 通知创建事件。
type NotificationCreatedEvent struct {
	NotificationID uint64    `json:"notification_id"`
	UserID         uint64    `json:"user_id"`
	Timestamp      time.Time `json:"timestamp"`
}

// NotificationReadEvent 通知已读事件。
type NotificationReadEvent struct {
	NotificationID uint64    `json:"notification_id"`
	UserID         uint64    `json:"user_id"`
	Timestamp      time.Time `json:"timestamp"`
}

// NotificationDeletedEvent 通知删除事件。
type NotificationDeletedEvent struct {
	NotificationID uint64    `json:"notification_id"`
	UserID         uint64    `json:"user_id"`
	Timestamp      time.Time `json:"timestamp"`
}

// NotificationTemplateCreatedEvent 通知模板创建事件。
type NotificationTemplateCreatedEvent struct {
	TemplateID uint64    `json:"template_id"`
	Code       string    `json:"code"`
	Timestamp  time.Time `json:"timestamp"`
}
