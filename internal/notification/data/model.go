package data

import (
	"time"

	"gorm.io/gorm"
)

<<<<<<< HEAD
// Notification 通知记录
type Notification struct {
	gorm.Model
	NotificationID string `gorm:"uniqueIndex;not null;comment:通知唯一ID" json:"notificationId"`
	UserID         uint64 `gorm:"index;not null;comment:用户ID" json:"userId"`
	Type           string `gorm:"not null;size:50;comment:通知类型 (SYSTEM, ORDER, MARKETING, INTERACTION)" json:"type"`
	Title          string `gorm:"not null;size:255;comment:通知标题" json:"title"`
	Content        string `gorm:"not null;type:text;comment:通知内容" json:"content"`
	IsRead         bool   `gorm:"not null;default:false;comment:是否已读" json:"isRead"`
}

// TableName 指定 Notification 的表名
=======
// Notification represents a notification record.
type Notification struct {
	gorm.Model
	NotificationID string    `gorm:"uniqueIndex;not null;comment:通知唯一ID" json:"notificationId"`
	UserID         uint64    `gorm:"index;not null;comment:用户ID" json:"userId"`
	Type           string    `gorm:"not null;size:50;comment:通知类型 (SYSTEM, ORDER, MARKETING, INTERACTION)" json:"type"`
	Title          string    `gorm:"not null;size:255;comment:通知标题" json:"title"`
	Content        string    `gorm:"not null;type:text;comment:通知内容" json:"content"`
	IsRead         bool      `gorm:"not null;default:false;comment:是否已读" json:"isRead"`
}

// TableName specifies the table name for Notification.
>>>>>>> 04d1270d593e17e866ec0ca4dad1f5d56021f07d
func (Notification) TableName() string {
	return "notifications"
}
