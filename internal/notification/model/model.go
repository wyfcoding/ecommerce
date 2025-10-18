package model

import (
	"time"

	"gorm.io/gorm"
)

// NotificationChannel 定义了通知的渠道类型
type NotificationChannel string

const (
	ChannelEmail NotificationChannel = "email"
	ChannelSMS   NotificationChannel = "sms"
	ChannelPush  NotificationChannel = "push"
)

// NotificationStatus 定义了通知发送的状态
type NotificationStatus int

const (
	StatusPending NotificationStatus = iota + 1 // 1: 待发送
	StatusSent                          // 2: 发送成功
	StatusFailed                        // 3: 发送失败
)

// NotificationLog 通知发送日志模型
// 记录每一次发送通知的尝试
type NotificationLog struct {
	ID              uint                `gorm:"primarykey" json:"id"`
	UserID          uint                `gorm:"index" json:"user_id"`                                // 接收用户ID
	Channel         NotificationChannel `gorm:"type:varchar(20);not null" json:"channel"`              // 发送渠道 (email, sms, push)
	Recipient       string              `gorm:"type:varchar(255);not null" json:"recipient"`            // 接收地址 (邮箱地址、手机号、设备token等)
	Subject         string              `gorm:"type:varchar(255)" json:"subject"`                      // 标题 (主要用于邮件)
	Content         string              `gorm:"type:text;not null" json:"content"`                     // 通知内容
	Status          NotificationStatus  `gorm:"not null;default:1" json:"status"`                      // 发送状态
	FailureReason   string              `gorm:"type:text" json:"failure_reason"`                       // 发送失败的原因
	SentAt          *time.Time          `json:"sent_at"`                                               // 发送成功的时间
	TemplateID      string              `gorm:"type:varchar(100);index" json:"template_id"`            // 使用的模板ID
	CorrelationID   string              `gorm:"type:varchar(100);index" json:"correlation_id"`         // 关联ID (例如，订单号、用户注册事件ID)
	CreatedAt       time.Time           `json:"created_at"`
}

// NotificationTemplate 通知模板模型
// 存储可复用的通知内容模板
type NotificationTemplate struct {
	ID        string    `gorm:"primary_key;type:varchar(100)" json:"id"` // 模板唯一ID (例如: "user_welcome_email")
	Channel   NotificationChannel `gorm:"type:varchar(20);not null" json:"channel"` // 模板适用的渠道
	Subject   string    `gorm:"type:varchar(255)" json:"subject"`       // 模板标题 (支持Go template语法)
	Body      string    `gorm:"type:text;not null" json:"body"`          // 模板内容 (支持Go template语法)
	Language  string    `gorm:"type:varchar(10);not null" json:"language"` // 语言 (例如: "en-US", "zh-CN")
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName 自定义表名
func (NotificationLog) TableName() string {
	return "notification_logs"
}

func (NotificationTemplate) TableName() string {
	return "notification_templates"
}
