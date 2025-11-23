package entity

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"gorm.io/gorm"
)

// NotificationType 通知类型
type NotificationType string

const (
	NotificationTypeSystem  NotificationType = "SYSTEM"  // 系统通知
	NotificationTypeOrder   NotificationType = "ORDER"   // 订单通知
	NotificationTypePayment NotificationType = "PAYMENT" // 支付通知
	NotificationTypePromo   NotificationType = "PROMO"   // 促销通知
)

// NotificationChannel 通知渠道
type NotificationChannel string

const (
	NotificationChannelApp   NotificationChannel = "APP"   // 应用内
	NotificationChannelSMS   NotificationChannel = "SMS"   // 短信
	NotificationChannelEmail NotificationChannel = "EMAIL" // 邮件
	NotificationChannelPush  NotificationChannel = "PUSH"  // 推送
)

// NotificationStatus 通知状态
type NotificationStatus int8

const (
	NotificationStatusUnread  NotificationStatus = 0 // 未读
	NotificationStatusRead    NotificationStatus = 1 // 已读
	NotificationStatusDeleted NotificationStatus = 2 // 已删除
)

// JSONMap defines a map that implements the sql.Scanner and driver.Valuer interfaces
type JSONMap map[string]interface{}

func (m JSONMap) Value() (driver.Value, error) {
	if m == nil {
		return nil, nil
	}
	return json.Marshal(m)
}

func (m *JSONMap) Scan(value interface{}) error {
	if value == nil {
		*m = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, m)
}

// StringArray defines a slice of strings that implements the sql.Scanner and driver.Valuer interfaces
type StringArray []string

func (a StringArray) Value() (driver.Value, error) {
	if a == nil {
		return nil, nil
	}
	return json.Marshal(a)
}

func (a *StringArray) Scan(value interface{}) error {
	if value == nil {
		*a = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, a)
}

// Notification 通知实体
type Notification struct {
	gorm.Model
	UserID    uint64              `gorm:"not null;index;comment:用户ID" json:"user_id"`
	NotifType NotificationType    `gorm:"type:varchar(32);not null;comment:通知类型" json:"notif_type"`
	Channel   NotificationChannel `gorm:"type:varchar(32);not null;comment:通知渠道" json:"channel"`
	Title     string              `gorm:"type:varchar(255);not null;comment:标题" json:"title"`
	Content   string              `gorm:"type:text;not null;comment:内容" json:"content"`
	Data      JSONMap             `gorm:"type:json;comment:扩展数据" json:"data"`
	Status    NotificationStatus  `gorm:"type:tinyint;not null;default:0;comment:状态" json:"status"`
	ReadAt    *time.Time          `gorm:"comment:阅读时间" json:"read_at"`
}

// NewNotification 创建通知
func NewNotification(userID uint64, notifType NotificationType, channel NotificationChannel, title, content string, data map[string]interface{}) *Notification {
	return &Notification{
		UserID:    userID,
		NotifType: notifType,
		Channel:   channel,
		Title:     title,
		Content:   content,
		Data:      data,
		Status:    NotificationStatusUnread,
	}
}

// MarkAsRead 标记为已读
func (n *Notification) MarkAsRead() {
	if n.Status == NotificationStatusUnread {
		n.Status = NotificationStatusRead
		now := time.Now()
		n.ReadAt = &now
	}
}

// NotificationTemplate 通知模板实体
type NotificationTemplate struct {
	gorm.Model
	Code      string              `gorm:"type:varchar(64);uniqueIndex;not null;comment:模板代码" json:"code"`
	Name      string              `gorm:"type:varchar(255);not null;comment:模板名称" json:"name"`
	NotifType NotificationType    `gorm:"type:varchar(32);not null;comment:通知类型" json:"notif_type"`
	Channel   NotificationChannel `gorm:"type:varchar(32);not null;comment:通知渠道" json:"channel"`
	Title     string              `gorm:"type:varchar(255);not null;comment:标题模板" json:"title"`
	Content   string              `gorm:"type:text;not null;comment:内容模板" json:"content"`
	Variables StringArray         `gorm:"type:json;comment:变量列表" json:"variables"`
	Enabled   bool                `gorm:"default:true;comment:是否启用" json:"enabled"`
}
