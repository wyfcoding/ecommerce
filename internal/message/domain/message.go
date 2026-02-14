package domain

import (
	"errors"
	"time"
)

var (
	ErrMessageNotFound       = errors.New("message not found")
	ErrMessageAlreadyRead    = errors.New("message already read")
	ErrMessageAlreadyDeleted = errors.New("message already deleted")
	ErrTemplateNotFound      = errors.New("template not found")
	ErrInvalidTemplate       = errors.New("invalid template")
)

type MessageType string

const (
	MessageTypeSystem       MessageType = "SYSTEM"
	MessageTypeOrder        MessageType = "ORDER"
	MessageTypePayment      MessageType = "PAYMENT"
	MessageTypePromotion    MessageType = "PROMOTION"
	MessageTypeActivity     MessageType = "ACTIVITY"
	MessageTypeNotification MessageType = "NOTIFICATION"
	MessageTypeAnnouncement MessageType = "ANNOUNCEMENT"
	MessageTypeAlert        MessageType = "ALERT"
)

type MessageStatus string

const (
	MessageStatusUnread   MessageStatus = "UNREAD"
	MessageStatusRead     MessageStatus = "READ"
	MessageStatusDeleted  MessageStatus = "DELETED"
	MessageStatusArchived MessageStatus = "ARCHIVED"
)

type MessagePriority string

const (
	MessagePriorityLow    MessagePriority = "LOW"
	MessagePriorityNormal MessagePriority = "NORMAL"
	MessagePriorityHigh   MessagePriority = "HIGH"
	MessagePriorityUrgent MessagePriority = "URGENT"
)

type UserMessage struct {
	ID           uint                 `json:"id"`
	CreatedAt    time.Time            `json:"created_at"`
	UpdatedAt    time.Time            `json:"updated_at"`
	UserID       uint64               `json:"user_id"`
	MessageID    uint64               `json:"message_id"`
	Type         MessageType          `json:"type"`
	Title        string               `json:"title"`
	Content      string               `json:"content"`
	Summary      string               `json:"summary"`
	Status       MessageStatus        `json:"status"`
	Priority     MessagePriority      `json:"priority"`
	Category     string               `json:"category"`
	SubCategory  string               `json:"sub_category"`
	SenderID     uint64               `json:"sender_id"`
	SenderName   string               `json:"sender_name"`
	SenderAvatar string               `json:"sender_avatar"`
	Link         string               `json:"link"`
	ImageURL     string               `json:"image_url"`
	Attachments  []*MessageAttachment `json:"attachments"`
	ExtraData    string               `json:"extra_data"`
	ReadAt       *time.Time           `json:"read_at"`
	DeletedAt    *time.Time           `json:"deleted_at"`
	ArchivedAt   *time.Time           `json:"archived_at"`
	ExpiresAt    *time.Time           `json:"expires_at"`
	PushedAt     *time.Time           `json:"pushed_at"`
	PushStatus   string               `json:"push_status"`
}

type MessageAttachment struct {
	ID        uint      `json:"id"`
	MessageID uint      `json:"message_id"`
	Name      string    `json:"name"`
	URL       string    `json:"url"`
	Size      int64     `json:"size"`
	Type      string    `json:"type"`
	CreatedAt time.Time `json:"created_at"`
}

type MessageTemplate struct {
	ID          uint        `json:"id"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
	Code        string      `json:"code"`
	Name        string      `json:"name"`
	Type        MessageType `json:"type"`
	Title       string      `json:"title"`
	Content     string      `json:"content"`
	Summary     string      `json:"summary"`
	Category    string      `json:"category"`
	Variables   []string    `json:"variables"`
	Enabled     bool        `json:"enabled"`
	Description string      `json:"description"`
}

type MessageBatch struct {
	ID           uint        `json:"id"`
	CreatedAt    time.Time   `json:"created_at"`
	UpdatedAt    time.Time   `json:"updated_at"`
	BatchNo      string      `json:"batch_no"`
	Name         string      `json:"name"`
	Type         MessageType `json:"type"`
	TemplateID   uint        `json:"template_id"`
	Title        string      `json:"title"`
	Content      string      `json:"content"`
	TargetType   string      `json:"target_type"`
	TargetCount  int         `json:"target_count"`
	SentCount    int         `json:"sent_count"`
	SuccessCount int         `json:"success_count"`
	FailCount    int         `json:"fail_count"`
	Status       string      `json:"status"`
	ScheduledAt  *time.Time  `json:"scheduled_at"`
	SentAt       *time.Time  `json:"sent_at"`
	CompletedAt  *time.Time  `json:"completed_at"`
}

type MessageCategory struct {
	ID          uint      `json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Code        string    `json:"code"`
	Name        string    `json:"name"`
	ParentCode  string    `json:"parent_code"`
	Icon        string    `json:"icon"`
	SortOrder   int       `json:"sort_order"`
	Enabled     bool      `json:"enabled"`
	Description string    `json:"description"`
}

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
	QuietHoursStart string    `json:"quiet_hours_start"`
	QuietHoursEnd   string    `json:"quiet_hours_end"`
}

type MessageStatistics struct {
	ID              uint        `json:"id"`
	Date            time.Time   `json:"date"`
	Type            MessageType `json:"type"`
	TotalSent       int64       `json:"total_sent"`
	TotalRead       int64       `json:"total_read"`
	TotalDeleted    int64       `json:"total_deleted"`
	ReadRate        float64     `json:"read_rate"`
	AvgReadTime     int64       `json:"avg_read_time"`
	PushSuccess     int64       `json:"push_success"`
	PushFail        int64       `json:"push_fail"`
	PushSuccessRate float64     `json:"push_success_rate"`
}

func NewUserMessage(userID uint64, msgType MessageType, title, content string) *UserMessage {
	return &UserMessage{
		UserID:      userID,
		Type:        msgType,
		Title:       title,
		Content:     content,
		Status:      MessageStatusUnread,
		Priority:    MessagePriorityNormal,
		Attachments: make([]*MessageAttachment, 0),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

func (m *UserMessage) SetSender(senderID uint64, name, avatar string) {
	m.SenderID = senderID
	m.SenderName = name
	m.SenderAvatar = avatar
	m.UpdatedAt = time.Now()
}

func (m *UserMessage) SetCategory(category, subCategory string) {
	m.Category = category
	m.SubCategory = subCategory
	m.UpdatedAt = time.Now()
}

func (m *UserMessage) SetLink(link string) {
	m.Link = link
	m.UpdatedAt = time.Now()
}

func (m *UserMessage) SetImage(url string) {
	m.ImageURL = url
	m.UpdatedAt = time.Now()
}

func (m *UserMessage) SetExtra(data string) {
	m.ExtraData = data
	m.UpdatedAt = time.Now()
}

func (m *UserMessage) SetPriority(priority MessagePriority) {
	m.Priority = priority
	m.UpdatedAt = time.Now()
}

func (m *UserMessage) SetExpiry(expiresAt time.Time) {
	m.ExpiresAt = &expiresAt
	m.UpdatedAt = time.Now()
}

func (m *UserMessage) AddAttachment(name, url string, size int64, fileType string) {
	attachment := &MessageAttachment{
		Name:      name,
		URL:       url,
		Size:      size,
		Type:      fileType,
		CreatedAt: time.Now(),
	}
	m.Attachments = append(m.Attachments, attachment)
	m.UpdatedAt = time.Now()
}

func (m *UserMessage) Read() error {
	if m.Status == MessageStatusRead {
		return ErrMessageAlreadyRead
	}
	m.Status = MessageStatusRead
	now := time.Now()
	m.ReadAt = &now
	m.UpdatedAt = now
	return nil
}

func (m *UserMessage) Delete() error {
	if m.Status == MessageStatusDeleted {
		return ErrMessageAlreadyDeleted
	}
	m.Status = MessageStatusDeleted
	now := time.Now()
	m.DeletedAt = &now
	m.UpdatedAt = now
	return nil
}

func (m *UserMessage) Archive() {
	m.Status = MessageStatusArchived
	now := time.Now()
	m.ArchivedAt = &now
	m.UpdatedAt = now
}

func (m *UserMessage) MarkPushed(status string) {
	m.PushStatus = status
	now := time.Now()
	m.PushedAt = &now
	m.UpdatedAt = now
}

func (m *UserMessage) IsExpired() bool {
	if m.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*m.ExpiresAt)
}

func (m *UserMessage) IsRead() bool {
	return m.Status == MessageStatusRead
}

func (m *UserMessage) IsDeleted() bool {
	return m.Status == MessageStatusDeleted
}

func NewMessageTemplate(code, name string, msgType MessageType, title, content string) *MessageTemplate {
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

func (t *MessageTemplate) AddVariable(variable string) {
	t.Variables = append(t.Variables, variable)
	t.UpdatedAt = time.Now()
}

func (t *MessageTemplate) Enable() {
	t.Enabled = true
	t.UpdatedAt = time.Now()
}

func (t *MessageTemplate) Disable() {
	t.Enabled = false
	t.UpdatedAt = time.Now()
}

func NewMessageBatch(name string, msgType MessageType, templateID uint) *MessageBatch {
	return &MessageBatch{
		BatchNo:    generateBatchNo(),
		Name:       name,
		Type:       msgType,
		TemplateID: templateID,
		Status:     "PENDING",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
}

func generateBatchNo() string {
	return "BCH" + time.Now().Format("20060102150405")
}

func (b *MessageBatch) SetTarget(targetType string, count int) {
	b.TargetType = targetType
	b.TargetCount = count
	b.UpdatedAt = time.Now()
}

func (b *MessageBatch) Schedule(scheduledAt time.Time) {
	b.ScheduledAt = &scheduledAt
	b.UpdatedAt = time.Now()
}

func (b *MessageBatch) Start() {
	b.Status = "SENDING"
	now := time.Now()
	b.SentAt = &now
	b.UpdatedAt = now
}

func (b *MessageBatch) RecordSent(success bool) {
	b.SentCount++
	if success {
		b.SuccessCount++
	} else {
		b.FailCount++
	}
	b.UpdatedAt = time.Now()
}

func (b *MessageBatch) Complete() {
	b.Status = "COMPLETED"
	now := time.Now()
	b.CompletedAt = &now
	b.UpdatedAt = now
}

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

func (s *MessageSetting) SetQuietHours(start, end string) {
	s.QuietHoursStart = start
	s.QuietHoursEnd = end
	s.UpdatedAt = time.Now()
}

func (s *MessageSetting) EnablePush(enabled bool) {
	s.PushEnabled = enabled
	s.UpdatedAt = time.Now()
}

func (s *MessageSetting) EnableEmail(enabled bool) {
	s.EmailEnabled = enabled
	s.UpdatedAt = time.Now()
}

func (s *MessageSetting) EnableSMS(enabled bool) {
	s.SMSEnabled = enabled
	s.UpdatedAt = time.Now()
}

func (s *MessageSetting) EnableInApp(enabled bool) {
	s.InAppEnabled = enabled
	s.UpdatedAt = time.Now()
}

type MessageRepository interface {
	Save(ctx any, message *UserMessage) error
	Update(ctx any, message *UserMessage) error
	FindByID(ctx any, id uint) (*UserMessage, error)
	FindByUserID(ctx any, userID uint64, status MessageStatus, limit, offset int) ([]*UserMessage, error)
	FindUnreadByUserID(ctx any, userID uint64) ([]*UserMessage, error)
	CountUnreadByUserID(ctx any, userID uint64) (int64, error)
	CountByUserID(ctx any, userID uint64, status MessageStatus) (int64, error)
	DeleteByUserID(ctx any, userID uint64) error
	MarkAllRead(ctx any, userID uint64) error
	DeleteExpiredMessages(ctx any) error

	SaveTemplate(ctx any, template *MessageTemplate) error
	FindTemplateByID(ctx any, id uint) (*MessageTemplate, error)
	FindTemplateByCode(ctx any, code string) (*MessageTemplate, error)
	FindEnabledTemplates(ctx any) ([]*MessageTemplate, error)

	SaveBatch(ctx any, batch *MessageBatch) error
	FindBatchByID(ctx any, id uint) (*MessageBatch, error)
	FindPendingBatches(ctx any) ([]*MessageBatch, error)

	SaveSetting(ctx any, setting *MessageSetting) error
	FindSettingByUserID(ctx any, userID uint64) (*MessageSetting, error)

	SaveStatistics(ctx any, stats *MessageStatistics) error
	FindStatisticsByDate(ctx any, date time.Time) ([]*MessageStatistics, error)
}

type MessageService interface {
	SendMessage(ctx any, userID uint64, msgType MessageType, title, content string, opts ...MessageOption) (*UserMessage, error)
	SendBatchMessages(ctx any, userIDs []uint64, msgType MessageType, title, content string) error
	SendFromTemplate(ctx any, userID uint64, templateCode string, variables map[string]string) (*UserMessage, error)
	GetUserMessages(ctx any, userID uint64, status MessageStatus, limit, offset int) ([]*UserMessage, error)
	GetUnreadMessages(ctx any, userID uint64) ([]*UserMessage, error)
	GetUnreadCount(ctx any, userID uint64) (int64, error)
	MarkAsRead(ctx any, messageID uint) error
	MarkAllAsRead(ctx any, userID uint64) error
	DeleteMessage(ctx any, messageID uint) error
	CreateTemplate(ctx any, template *MessageTemplate) error
	GetTemplate(ctx any, code string) (*MessageTemplate, error)
	GetUserSettings(ctx any, userID uint64) (*MessageSetting, error)
	UpdateUserSettings(ctx any, userID uint64, settings map[string]bool) error
}

type MessageOption func(*UserMessage)

func WithSender(senderID uint64, name, avatar string) MessageOption {
	return func(m *UserMessage) {
		m.SetSender(senderID, name, avatar)
	}
}

func WithCategory(category, subCategory string) MessageOption {
	return func(m *UserMessage) {
		m.SetCategory(category, subCategory)
	}
}

func WithLink(link string) MessageOption {
	return func(m *UserMessage) {
		m.SetLink(link)
	}
}

func WithImage(url string) MessageOption {
	return func(m *UserMessage) {
		m.SetImage(url)
	}
}

func WithPriority(priority MessagePriority) MessageOption {
	return func(m *UserMessage) {
		m.SetPriority(priority)
	}
}

func WithExpiry(expiresAt time.Time) MessageOption {
	return func(m *UserMessage) {
		m.SetExpiry(expiresAt)
	}
}

func WithExtra(data string) MessageOption {
	return func(m *UserMessage) {
		m.SetExtra(data)
	}
}
