package domain

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

// Sender 定义了通知发送器的通用接口
// 目标为站内信/邮件/短信/Webhook 等。
type Sender interface {
	Send(ctx context.Context, target, subject, content string) error
}

// NotificationType 定义了通知的类型。
type NotificationType string

const (
	NotificationTypeSystem  NotificationType = "SYSTEM"  // 系统通知，例如账户安全提示。
	NotificationTypeOrder   NotificationType = "ORDER"   // 订单通知，例如订单状态变更。
	NotificationTypePayment NotificationType = "PAYMENT" // 支付通知，例如支付成功。
	NotificationTypePromo   NotificationType = "PROMO"   // 促销通知，例如优惠活动。
)

// NotificationChannel 定义了通知发送的渠道。
type NotificationChannel string

const (
	NotificationChannelApp     NotificationChannel = "APP"     // 站内信/应用程序内通知。
	NotificationChannelSMS     NotificationChannel = "SMS"     // 短信通知。
	NotificationChannelEmail   NotificationChannel = "EMAIL"   // 邮件通知。
	NotificationChannelPush    NotificationChannel = "PUSH"    // 推送通知（例如，App Push）。
	NotificationChannelWebhook NotificationChannel = "WEBHOOK" // Webhook通知。
)

// NotificationStatus 定义了通知的阅读状态。
type NotificationStatus int8

const (
	NotificationStatusUnread    NotificationStatus = 0 // 未读。
	NotificationStatusRead      NotificationStatus = 1 // 已读。
	NotificationStatusDelivered NotificationStatus = 3 // 已送达。
	NotificationStatusDeleted   NotificationStatus = 2 // 已删除。
)

// JSONMap 定义了一个map类型，实现了 sql.Scanner 和 driver.Valuer 接口，
// 允许将 Go 的 map[string]any 类型作为 JSON 字符串存储到数据库。
type JSONMap map[string]any

// Value 实现 driver.Valuer 接口，将 JSONMap 转换为数据库可以存储的值（JSON字节数组）。
func (m JSONMap) Value() (driver.Value, error) {
	if m == nil {
		return nil, nil
	}
	return json.Marshal(m)
}

// Scan 实现 sql.Scanner 接口，从数据库读取值并转换为 JSONMap。
func (m *JSONMap) Scan(value any) error {
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

// StringArray 定义了一个字符串切片类型，实现了 sql.Scanner 和 driver.Valuer 接口。
type StringArray []string

// Value 实现 driver.Valuer 接口，将 StringArray 转换为数据库可以存储的值（JSON字节数组）。
func (a StringArray) Value() (driver.Value, error) {
	if a == nil {
		return nil, nil
	}
	return json.Marshal(a)
}

// Scan 实现 sql.Scanner 接口，从数据库读取值并转换为 StringArray。
func (a *StringArray) Scan(value any) error {
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

// Notification 实体代表一条发送给用户的通知。
type Notification struct {
	ID          uint64              `json:"id"`
	CreatedAt   time.Time           `json:"created_at"`
	UpdatedAt   time.Time           `json:"updated_at"`
	UserID      uint64              `json:"user_id"`
	Type        NotificationType    `json:"type"`
	Channel     NotificationChannel `json:"channel"`
	Title       string              `json:"title"`
	Content     string              `json:"content"`
	Data        JSONMap             `json:"data"`
	Status      NotificationStatus  `json:"status"`
	ReadAt      *time.Time          `json:"read_at"`
	DeliveredAt *time.Time          `json:"delivered_at"`
	NotifType   NotificationType    `json:"notif_type" gorm:"column:notif_type"`
	Metadata    JSONMap             `json:"metadata"`
}

// NewNotification 创建并返回一个新的 Notification 实体实例。
func NewNotification(userID uint64, notifType NotificationType, channel NotificationChannel, title, content string, data map[string]any) *Notification {
	return &Notification{
		UserID:  userID,
		Type:    notifType,
		Channel: channel,
		Title:   title,
		Content: content,
		Data:    data,
		Status:  NotificationStatusUnread,
	}
}

// MarkAsRead 标记通知为已读。
func (n *Notification) MarkAsRead() {
	if n.Status == NotificationStatusUnread {
		n.Status = NotificationStatusRead
		now := time.Now()
		n.ReadAt = &now
	}
}

// NotificationTemplate 实体代表一个通知模板。
type NotificationTemplate struct {
	ID        uint64              `json:"id"`
	CreatedAt time.Time           `json:"created_at"`
	UpdatedAt time.Time           `json:"updated_at"`
	Code      string              `json:"code"`
	Name      string              `json:"name"`
	Type      NotificationType    `json:"type"`
	Channel   NotificationChannel `json:"channel"`
	Title     string              `json:"title"`
	Content   string              `json:"content"`
	Variables StringArray         `json:"variables"`
	NotifType NotificationType    `json:"notif_type" gorm:"column:notif_type"`
	Enabled   bool                `json:"enabled"`
}
