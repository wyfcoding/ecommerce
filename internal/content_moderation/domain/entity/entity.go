package entity

import (
	"time"

	"gorm.io/gorm"
)

// ContentType 内容类型
type ContentType string

const (
	ContentTypeText  ContentType = "TEXT"  // 文本
	ContentTypeImage ContentType = "IMAGE" // 图片
	ContentTypeVideo ContentType = "VIDEO" // 视频
	ContentTypeAudio ContentType = "AUDIO" // 音频
)

// ModerationStatus 审核状态
type ModerationStatus int8

const (
	ModerationStatusPending  ModerationStatus = 0 // 待审核
	ModerationStatusApproved ModerationStatus = 1 // 通过
	ModerationStatusRejected ModerationStatus = 2 // 拒绝
)

// ModerationRecord 审核记录实体
type ModerationRecord struct {
	gorm.Model
	ContentType  ContentType      `gorm:"type:varchar(32);not null;comment:内容类型" json:"content_type"`
	ContentID    uint64           `gorm:"not null;index;comment:内容ID" json:"content_id"`
	Content      string           `gorm:"type:text;comment:内容" json:"content"`
	UserID       uint64           `gorm:"not null;index;comment:用户ID" json:"user_id"`
	Status       ModerationStatus `gorm:"default:0;comment:状态" json:"status"`
	AIScore      float64          `gorm:"type:decimal(5,4);comment:AI评分" json:"ai_score"`
	AITags       []string         `gorm:"type:json;serializer:json;comment:AI标签" json:"ai_tags"`
	RejectReason string           `gorm:"type:varchar(255);comment:拒绝原因" json:"reject_reason"`
	ModeratorID  uint64           `gorm:"comment:审核人ID" json:"moderator_id"`
	ModeratedAt  *time.Time       `gorm:"comment:审核时间" json:"moderated_at"`
}

// SensitiveWord 敏感词实体
type SensitiveWord struct {
	gorm.Model
	Word     string `gorm:"type:varchar(64);uniqueIndex;not null;comment:敏感词" json:"word"`
	Category string `gorm:"type:varchar(32);not null;comment:分类" json:"category"`
	Level    int8   `gorm:"default:1;comment:等级" json:"level"`
	Enabled  bool   `gorm:"default:true;comment:是否启用" json:"enabled"`
}

// NewModerationRecord 创建审核记录
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

// SetAIResult 设置AI审核结果
func (m *ModerationRecord) SetAIResult(score float64, tags []string) {
	m.AIScore = score
	m.AITags = tags

	// 自动审核
	if score < 0.3 {
		m.AutoApprove()
	} else if score > 0.8 {
		m.AutoReject("AI检测到违规内容")
	}
}

// Approve 审核通过
func (m *ModerationRecord) Approve(moderatorID uint64) {
	m.Status = ModerationStatusApproved
	m.ModeratorID = moderatorID
	now := time.Now()
	m.ModeratedAt = &now
}

// Reject 审核拒绝
func (m *ModerationRecord) Reject(moderatorID uint64, reason string) {
	m.Status = ModerationStatusRejected
	m.ModeratorID = moderatorID
	m.RejectReason = reason
	now := time.Now()
	m.ModeratedAt = &now
}

// AutoApprove 自动通过
func (m *ModerationRecord) AutoApprove() {
	m.Status = ModerationStatusApproved
	now := time.Now()
	m.ModeratedAt = &now
}

// AutoReject 自动拒绝
func (m *ModerationRecord) AutoReject(reason string) {
	m.Status = ModerationStatusRejected
	m.RejectReason = reason
	now := time.Now()
	m.ModeratedAt = &now
}

// NewSensitiveWord 创建敏感词
func NewSensitiveWord(word, category string, level int8) *SensitiveWord {
	return &SensitiveWord{
		Word:     word,
		Category: category,
		Level:    level,
		Enabled:  true,
	}
}

// Enable 启用
func (s *SensitiveWord) Enable() {
	s.Enabled = true
}

// Disable 禁用
func (s *SensitiveWord) Disable() {
	s.Enabled = false
}
