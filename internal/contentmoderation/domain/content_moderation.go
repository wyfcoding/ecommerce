package domain

import "time"

// ContentType 定义了待审核内容的类型。
type ContentType string

const (
	ContentTypeText  ContentType = "TEXT"
	ContentTypeImage ContentType = "IMAGE"
	ContentTypeVideo ContentType = "VIDEO"
	ContentTypeAudio ContentType = "AUDIO"
)

// ModerationStatus 定义了内容审核记录的状态。
type ModerationStatus int8

const (
	ModerationStatusPending  ModerationStatus = 0 // 待审核。
	ModerationStatusApproved ModerationStatus = 1 // 通过。
	ModerationStatusRejected ModerationStatus = 2 // 拒绝。
)

// ModerationRecord 实体代表一条内容审核记录。
type ModerationRecord struct {
	ID           uint             `json:"id"`
	CreatedAt    time.Time        `json:"created_at"`
	UpdatedAt    time.Time        `json:"updated_at"`
	ContentType  ContentType      `json:"content_type"`
	ContentID    uint64           `json:"content_id"`
	Content      string           `json:"content"`
	UserID       uint64           `json:"user_id"`
	Status       ModerationStatus `json:"status"`
	AIScore      float64          `json:"ai_score"`
	AITags       []string         `json:"ai_tags"`
	RejectReason string           `json:"reject_reason"`
	ModeratorID  uint64           `json:"moderator_id"`
	ModeratedAt  *time.Time       `json:"moderated_at"`
}

// SensitiveWord 实体代表一个敏感词。
type SensitiveWord struct {
	ID        uint      `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Word      string    `json:"word"`
	Category  string    `json:"category"`
	Level     int8      `json:"level"`
	Enabled   bool      `json:"enabled"`
}

// NewModerationRecord 创建并返回一个新的 ModerationRecord 实体实例。
func NewModerationRecord(contentType ContentType, contentID uint64, content string, userID uint64) *ModerationRecord {
	return &ModerationRecord{
		ContentType: contentType,
		ContentID:   contentID,
		Content:     content,
		UserID:      userID,
		Status:      ModerationStatusPending,
		AITags:      []string{},
	}
}

// SetAIResult 设置AI审核结果，并根据AI评分自动进行初步审核。
func (m *ModerationRecord) SetAIResult(score float64, tags []string) {
	m.AIScore = score
	m.AITags = tags
	if score < 0.3 {
		m.AutoApprove()
	} else if score > 0.8 {
		m.AutoReject("AI检测到违规内容")
	}
}

// Approve 批准审核记录。
func (m *ModerationRecord) Approve(moderatorID uint64) {
	m.Status = ModerationStatusApproved
	m.ModeratorID = moderatorID
	now := time.Now()
	m.ModeratedAt = &now
}

// Reject 拒绝审核记录。
func (m *ModerationRecord) Reject(moderatorID uint64, reason string) {
	m.Status = ModerationStatusRejected
	m.ModeratorID = moderatorID
	m.RejectReason = reason
	now := time.Now()
	m.ModeratedAt = &now
}

// AutoApprove 自动批准审核记录。
func (m *ModerationRecord) AutoApprove() {
	m.Status = ModerationStatusApproved
	now := time.Now()
	m.ModeratedAt = &now
}

// AutoReject 自动拒绝审核记录。
func (m *ModerationRecord) AutoReject(reason string) {
	m.Status = ModerationStatusRejected
	m.RejectReason = reason
	now := time.Now()
	m.ModeratedAt = &now
}

// NewSensitiveWord 创建并返回一个新的 SensitiveWord 实体实例。
func NewSensitiveWord(word, category string, level int8) *SensitiveWord {
	return &SensitiveWord{
		Word:     word,
		Category: category,
		Level:    level,
		Enabled:  true,
	}
}

// Enable 启用敏感词。
func (s *SensitiveWord) Enable() {
	s.Enabled = true
}

// Disable 禁用敏感词。
func (s *SensitiveWord) Disable() {
	s.Enabled = false
}
