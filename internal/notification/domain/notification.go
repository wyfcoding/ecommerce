package domain

import (
	"context"
	"database/sql/driver" // 导入数据库驱动接口。
	"encoding/json"       // 导入JSON编码/解码库。
	"errors"              // 导入标准错误处理库。
	"time"                // 导入时间包。

	"gorm.io/gorm" // 导入GORM库。
)

// Sender 定义了通知发送器的通用接口
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
	NotificationChannelApp   NotificationChannel = "APP"   // 站内信/应用程序内通知。
	NotificationChannelSMS   NotificationChannel = "SMS"   // 短信通知。
	NotificationChannelEmail NotificationChannel = "EMAIL" // 邮件通知。
	NotificationChannelPush  NotificationChannel = "PUSH"  // 推送通知（例如，App Push）。
)

// NotificationStatus 定义了通知的阅读状态。
type NotificationStatus int8

const (
	NotificationStatusUnread  NotificationStatus = 0 // 未读。
	NotificationStatusRead    NotificationStatus = 1 // 已读。
	NotificationStatusDeleted NotificationStatus = 2 // 已删除。
)

// JSONMap 定义了一个map类型，实现了 sql.Scanner 和 driver.Valuer 接口，
// 允许GORM将Go的map[string]interface{}类型作为JSON字符串存储到数据库，并从数据库读取。
type JSONMap map[string]any

// Value 实现 driver.Valuer 接口，将 JSONMap 转换为数据库可以存储的值（JSON字节数组）。
func (m JSONMap) Value() (driver.Value, error) {
	if m == nil {
		return nil, nil
	}
	return json.Marshal(m) // 将map编码为JSON字节数组。
}

// Scan 实现 sql.Scanner 接口，从数据库读取值并转换为 JSONMap。
func (m *JSONMap) Scan(value any) error {
	if value == nil {
		*m = nil
		return nil
	}
	bytes, ok := value.([]byte) // 期望数据库返回字节数组。
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, m) // 将JSON字节数组解码为map。
}

// StringArray 定义了一个字符串切片类型，实现了 sql.Scanner 和 driver.Valuer 接口，
// 允许GORM将Go的 []string 类型作为JSON字符串存储到数据库，并从数据库读取。
type StringArray []string

// Value 实现 driver.Valuer 接口，将 StringArray 转换为数据库可以存储的值（JSON字节数组）。
func (a StringArray) Value() (driver.Value, error) {
	if a == nil {
		return nil, nil
	}
	return json.Marshal(a) // 将切片编码为JSON字节数组。
}

// Scan 实现 sql.Scanner 接口，从数据库读取值并转换为 StringArray。
func (a *StringArray) Scan(value any) error {
	if value == nil {
		*a = nil
		return nil
	}
	bytes, ok := value.([]byte) // 期望数据库返回字节数组。
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, a) // 将JSON字节数组解码为切片。
}

// Notification 实体代表一条发送给用户的通知。
// 它包含了通知的接收者、类型、渠道、内容和阅读状态等。
type Notification struct {
	gorm.Model                     // 嵌入gorm.Model，包含ID, CreatedAt, UpdatedAt, DeletedAt等通用字段。
	UserID     uint64              `gorm:"not null;index;comment:用户ID" json:"user_id"`               // 接收通知的用户ID，索引字段。
	NotifType  NotificationType    `gorm:"type:varchar(32);not null;comment:通知类型" json:"notif_type"` // 通知类型。
	Channel    NotificationChannel `gorm:"type:varchar(32);not null;comment:通知渠道" json:"channel"`    // 通知发送渠道。
	Title      string              `gorm:"type:varchar(255);not null;comment:标题" json:"title"`       // 通知标题。
	Content    string              `gorm:"type:text;not null;comment:内容" json:"content"`             // 通知内容。
	Data       JSONMap             `gorm:"type:json;comment:扩展数据" json:"data"`                       // 附加数据，存储为JSON。
	Status     NotificationStatus  `gorm:"type:tinyint;not null;default:0;comment:状态" json:"status"` // 通知状态，默认为未读。
	ReadAt     *time.Time          `gorm:"comment:阅读时间" json:"read_at"`                              // 通知被阅读的时间。
}

// NewNotification 创建并返回一个新的 Notification 实体实例。
// userID: 接收通知的用户ID。
// notifType: 通知类型。
// channel: 通知渠道。
// title: 标题。
// content: 内容。
// data: 附加数据。
func NewNotification(userID uint64, notifType NotificationType, channel NotificationChannel, title, content string, data map[string]any) *Notification {
	return &Notification{
		UserID:    userID,
		NotifType: notifType,
		Channel:   channel,
		Title:     title,
		Content:   content,
		Data:      data,
		Status:    NotificationStatusUnread, // 新创建的通知默认为未读状态。
	}
}

// MarkAsRead 标记通知为已读。
func (n *Notification) MarkAsRead() {
	if n.Status == NotificationStatusUnread { // 只有未读通知才能标记为已读。
		n.Status = NotificationStatusRead // 状态更新为已读。
		now := time.Now()
		n.ReadAt = &now // 记录阅读时间。
	}
}

// NotificationTemplate 实体代表一个通知模板。
// 它包含了模板的代码、名称、类型、渠道以及标题和内容模板，支持变量替换。
type NotificationTemplate struct {
	gorm.Model                     // 嵌入gorm.Model。
	Code       string              `gorm:"type:varchar(64);uniqueIndex;not null;comment:模板代码" json:"code"` // 模板代码，唯一索引，不允许为空。
	Name       string              `gorm:"type:varchar(255);not null;comment:模板名称" json:"name"`            // 模板名称。
	NotifType  NotificationType    `gorm:"type:varchar(32);not null;comment:通知类型" json:"notif_type"`       // 模板对应的通知类型。
	Channel    NotificationChannel `gorm:"type:varchar(32);not null;comment:通知渠道" json:"channel"`          // 模板对应的通知渠道。
	Title      string              `gorm:"type:varchar(255);not null;comment:标题模板" json:"title"`           // 通知标题的模板字符串。
	Content    string              `gorm:"type:text;not null;comment:内容模板" json:"content"`                 // 通知内容的模板字符串。
	Variables  StringArray         `gorm:"type:json;comment:变量列表" json:"variables"`                        // 模板中使用的变量名称列表，存储为JSON。
	Enabled    bool                `gorm:"default:true;comment:是否启用" json:"enabled"`                       // 模板是否启用。
}
