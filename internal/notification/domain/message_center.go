// 生成摘要：
// - 从 internal/message 服务合并而来，补充站内信、消息模板、批量消息、消息设置等能力
// - 与 notification.go 互补：notification 负责通知发送渠道，message_center 负责用户站内信管理
// - 新增 UserMessage（用户站内信）、MessageTemplate（消息模板）、
//   MessageBatch（批量消息）、MessageSetting（用户消息偏好设置）等领域对象
// - 新增 MessageCenterRepository 仓储接口

package domain

import (
	"context"
	"errors"
	"time"
)

var (
	// ErrMessageNotFound 站内信不存在
	ErrMessageNotFound = errors.New("message not found")
	// ErrMessageAlreadyRead 站内信已读
	ErrMessageAlreadyRead = errors.New("message already read")
	// ErrMessageAlreadyDeleted 站内信已删除
	ErrMessageAlreadyDeleted = errors.New("message already deleted")
	// ErrMsgTemplateNotFound 消息模板不存在
	ErrMsgTemplateNotFound = errors.New("message template not found")
)

// InboxMessageType 站内信消息类型
type InboxMessageType string

const (
	InboxMessageTypeSystem       InboxMessageType = "SYSTEM"
	InboxMessageTypeOrder        InboxMessageType = "ORDER"
	InboxMessageTypePayment      InboxMessageType = "PAYMENT"
	InboxMessageTypePromotion    InboxMessageType = "PROMOTION"
	InboxMessageTypeActivity     InboxMessageType = "ACTIVITY"
	InboxMessageTypeNotification InboxMessageType = "NOTIFICATION"
	InboxMessageTypeAnnouncement InboxMessageType = "ANNOUNCEMENT"
	InboxMessageTypeAlert        InboxMessageType = "ALERT"
)

// InboxMessageStatus 站内信消息状态
type InboxMessageStatus string

const (
	InboxMessageStatusUnread   InboxMessageStatus = "UNREAD"
	InboxMessageStatusRead     InboxMessageStatus = "READ"
	InboxMessageStatusDeleted  InboxMessageStatus = "DELETED"
	InboxMessageStatusArchived InboxMessageStatus = "ARCHIVED"
)

// InboxMessagePriority 站内信消息优先级
type InboxMessagePriority string

const (
	InboxMessagePriorityLow    InboxMessagePriority = "LOW"
	InboxMessagePriorityNormal InboxMessagePriority = "NORMAL"
	InboxMessagePriorityHigh   InboxMessagePriority = "HIGH"
	InboxMessagePriorityUrgent InboxMessagePriority = "URGENT"
)

// UserMessage 用户站内信实体
// 表示发送给特定用户的一条站内消息
type UserMessage struct {
	ID           uint                 `json:"id"`
	CreatedAt    time.Time            `json:"created_at"`
	UpdatedAt    time.Time            `json:"updated_at"`
	UserID       uint64               `json:"user_id"`
	MessageID    uint64               `json:"message_id"`
	Type         InboxMessageType     `json:"type"`
	Title        string               `json:"title"`
	Content      string               `json:"content"`
	Summary      string               `json:"summary"`
	Status       InboxMessageStatus   `json:"status"`
	Priority     InboxMessagePriority `json:"priority"`
	Category     string               `json:"category"`
	SubCategory  string               `json:"sub_category"`
	SenderID     uint64               `json:"sender_id"`
	SenderName   string               `json:"sender_name"`
	SenderAvatar string               `json:"sender_avatar"`
	Link         string               `json:"link"`
	ImageURL     string               `json:"image_url"`
	ExtraData    string               `json:"extra_data"`
	ReadAt       *time.Time           `json:"read_at"`
	DeletedAt    *time.Time           `json:"deleted_at"`
	ArchivedAt   *time.Time           `json:"archived_at"`
	ExpiresAt    *time.Time           `json:"expires_at"`
	PushedAt     *time.Time           `json:"pushed_at"`
	PushStatus   string               `json:"push_status"`
}

// MessageTemplate 消息模板
// 用于标准化消息内容，支持变量替换
type MessageTemplate struct {
	ID          uint             `json:"id"`
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at"`
	Code        string           `json:"code"`
	Name        string           `json:"name"`
	Type        InboxMessageType `json:"type"`
	Title       string           `json:"title"`
	Content     string           `json:"content"`
	Summary     string           `json:"summary"`
	Category    string           `json:"category"`
	Variables   []string         `json:"variables"`
	Enabled     bool             `json:"enabled"`
	Description string           `json:"description"`
}

// MessageBatch 批量消息任务
// 用于批量向多个用户发送消息
type MessageBatch struct {
	ID           uint             `json:"id"`
	CreatedAt    time.Time        `json:"created_at"`
	UpdatedAt    time.Time        `json:"updated_at"`
	BatchNo      string           `json:"batch_no"`
	Name         string           `json:"name"`
	Type         InboxMessageType `json:"type"`
	TemplateID   uint             `json:"template_id"`
	Title        string           `json:"title"`
	Content      string           `json:"content"`
	TargetType   string           `json:"target_type"` // 目标类型（ALL/GROUP/CUSTOM）
	TargetCount  int              `json:"target_count"`
	SentCount    int              `json:"sent_count"`
	SuccessCount int              `json:"success_count"`
	FailCount    int              `json:"fail_count"`
	Status       string           `json:"status"`
	ScheduledAt  *time.Time       `json:"scheduled_at"`
	SentAt       *time.Time       `json:"sent_at"`
	CompletedAt  *time.Time       `json:"completed_at"`
}

// MessageSetting 用户消息偏好设置
// 控制各渠道的消息推送开关和免打扰时段
type MessageSetting struct {
	ID              uint      `json:"id"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	UserID          uint64    `json:"user_id"`
	Category        string    `json:"category"`
	PushEnabled     bool      `json:"push_enabled"`
	EmailEnabled    bool      `json:"email_enabled"`
	SMSEnabled      bool      `json:"sms_enabled"`
	InAppEnabled    bool      `json:"in_app_enabled"`
	QuietHoursStart string    `json:"quiet_hours_start"` // 免打扰开始时间 HH:MM
	QuietHoursEnd   string    `json:"quiet_hours_end"`   // 免打扰结束时间 HH:MM
}

// MessageStatistics 消息统计
// 按日期和类型统计消息发送/阅读情况
type MessageStatistics struct {
	ID              uint             `json:"id"`
	Date            time.Time        `json:"date"`
	Type            InboxMessageType `json:"type"`
	TotalSent       int64            `json:"total_sent"`
	TotalRead       int64            `json:"total_read"`
	TotalDeleted    int64            `json:"total_deleted"`
	ReadRate        float64          `json:"read_rate"`
	AvgReadTime     int64            `json:"avg_read_time"`
	PushSuccess     int64            `json:"push_success"`
	PushFail        int64            `json:"push_fail"`
	PushSuccessRate float64          `json:"push_success_rate"`
}

// NewUserMessage 创建新的用户站内信
func NewUserMessage(userID uint64, msgType InboxMessageType, title, content string) *UserMessage {
	return &UserMessage{
		UserID:    userID,
		Type:      msgType,
		Title:     title,
		Content:   content,
		Status:    InboxMessageStatusUnread,
		Priority:  InboxMessagePriorityNormal,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// SetSender 设置消息发送者信息
func (m *UserMessage) SetSender(senderID uint64, name, avatar string) {
	m.SenderID = senderID
	m.SenderName = name
	m.SenderAvatar = avatar
	m.UpdatedAt = time.Now()
}

// SetCategory 设置消息分类
func (m *UserMessage) SetCategory(category, subCategory string) {
	m.Category = category
	m.SubCategory = subCategory
	m.UpdatedAt = time.Now()
}

// SetPriority 设置消息优先级
func (m *UserMessage) SetPriority(priority InboxMessagePriority) {
	m.Priority = priority
	m.UpdatedAt = time.Now()
}

// SetExpiry 设置消息过期时间
func (m *UserMessage) SetExpiry(expiresAt time.Time) {
	m.ExpiresAt = &expiresAt
	m.UpdatedAt = time.Now()
}

// Read 标记消息为已读
func (m *UserMessage) Read() error {
	if m.Status == InboxMessageStatusRead {
		return ErrMessageAlreadyRead
	}
	m.Status = InboxMessageStatusRead
	now := time.Now()
	m.ReadAt = &now
	m.UpdatedAt = now
	return nil
}

// Delete 标记消息为已删除
func (m *UserMessage) Delete() error {
	if m.Status == InboxMessageStatusDeleted {
		return ErrMessageAlreadyDeleted
	}
	m.Status = InboxMessageStatusDeleted
	now := time.Now()
	m.DeletedAt = &now
	m.UpdatedAt = now
	return nil
}

// Archive 归档消息
func (m *UserMessage) Archive() {
	m.Status = InboxMessageStatusArchived
	now := time.Now()
	m.ArchivedAt = &now
	m.UpdatedAt = now
}

// MarkPushed 标记消息已推送
func (m *UserMessage) MarkPushed(status string) {
	m.PushStatus = status
	now := time.Now()
	m.PushedAt = &now
	m.UpdatedAt = now
}

// IsExpired 检查消息是否已过期
func (m *UserMessage) IsExpired() bool {
	if m.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*m.ExpiresAt)
}

// NewMessageTemplate 创建新的消息模板
func NewMessageTemplate(code, name string, msgType InboxMessageType, title, content string) *MessageTemplate {
	return &MessageTemplate{
		Code:      code,
		Name:      name,
		Type:      msgType,
		Title:     title,
		Content:   content,
		Variables: make([]string, 0),
		Enabled:   true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// NewMessageBatch 创建新的批量消息任务
func NewMessageBatch(name string, msgType InboxMessageType, templateID uint) *MessageBatch {
	return &MessageBatch{
		BatchNo:    "BCH" + time.Now().Format("20060102150405"),
		Name:       name,
		Type:       msgType,
		TemplateID: templateID,
		Status:     "PENDING",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
}

// Start 开始发送批量消息
func (b *MessageBatch) Start() {
	b.Status = "SENDING"
	now := time.Now()
	b.SentAt = &now
	b.UpdatedAt = now
}

// RecordSent 记录单条消息发送结果
func (b *MessageBatch) RecordSent(success bool) {
	b.SentCount++
	if success {
		b.SuccessCount++
	} else {
		b.FailCount++
	}
	b.UpdatedAt = time.Now()
}

// Complete 完成批量发送
func (b *MessageBatch) Complete() {
	b.Status = "COMPLETED"
	now := time.Now()
	b.CompletedAt = &now
	b.UpdatedAt = now
}

// NewMessageSetting 创建用户默认消息设置
func NewMessageSetting(userID uint64) *MessageSetting {
	return &MessageSetting{
		UserID:       userID,
		PushEnabled:  true,
		EmailEnabled: true,
		SMSEnabled:   true,
		InAppEnabled: true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
}

// SetQuietHours 设置免打扰时段
func (s *MessageSetting) SetQuietHours(start, end string) {
	s.QuietHoursStart = start
	s.QuietHoursEnd = end
	s.UpdatedAt = time.Now()
}

// MessageCenterRepository 消息中心仓储接口
type MessageCenterRepository interface {
	// SaveMessage 保存站内信
	SaveMessage(ctx context.Context, message *UserMessage) error
	// UpdateMessage 更新站内信
	UpdateMessage(ctx context.Context, message *UserMessage) error
	// FindMessageByID 根据 ID 查找站内信
	FindMessageByID(ctx context.Context, id uint) (*UserMessage, error)
	// FindByUserID 根据用户 ID 查找站内信
	FindByUserID(ctx context.Context, userID uint64, status InboxMessageStatus, limit, offset int) ([]*UserMessage, error)
	// CountUnread 统计未读消息数
	CountUnread(ctx context.Context, userID uint64) (int64, error)
	// MarkAllRead 全部标记已读
	MarkAllRead(ctx context.Context, userID uint64) error
	// DeleteExpired 删除过期消息
	DeleteExpired(ctx context.Context) error

	// SaveTemplate 保存消息模板
	SaveTemplate(ctx context.Context, template *MessageTemplate) error
	// FindTemplateByCode 根据编码查找模板
	FindTemplateByCode(ctx context.Context, code string) (*MessageTemplate, error)

	// SaveBatch 保存批量消息任务
	SaveBatch(ctx context.Context, batch *MessageBatch) error
	// FindPendingBatches 查找待发送的批量任务
	FindPendingBatches(ctx context.Context) ([]*MessageBatch, error)

	// SaveSetting 保存用户消息设置
	SaveSetting(ctx context.Context, setting *MessageSetting) error
	// FindSettingByUserID 根据用户 ID 查找消息设置
	FindSettingByUserID(ctx context.Context, userID uint64) (*MessageSetting, error)

	// SaveStatistics 保存消息统计
	SaveStatistics(ctx context.Context, stats *MessageStatistics) error
}
