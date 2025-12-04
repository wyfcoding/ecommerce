package entity

import (
	"time"

	"gorm.io/gorm" // 导入GORM库。
)

// ContentType 定义了待审核内容的类型。
type ContentType string

const (
	ContentTypeText  ContentType = "TEXT"  // 文本内容，例如用户评论、商品描述。
	ContentTypeImage ContentType = "IMAGE" // 图片内容，例如用户上传的商品图片、头像。
	ContentTypeVideo ContentType = "VIDEO" // 视频内容。
	ContentTypeAudio ContentType = "AUDIO" // 音频内容。
)

// ModerationStatus 定义了内容审核记录的状态。
type ModerationStatus int8

const (
	ModerationStatusPending  ModerationStatus = 0 // 待审核：内容已提交，等待AI或人工审核。
	ModerationStatusApproved ModerationStatus = 1 // 通过：内容审核通过，可以发布。
	ModerationStatusRejected ModerationStatus = 2 // 拒绝：内容审核未通过，不能发布。
)

// ModerationRecord 实体代表一条内容审核记录。
// 它包含了待审核的内容、审核状态、AI审核结果和人工审核结果等信息。
type ModerationRecord struct {
	gorm.Model                    // 嵌入gorm.Model，包含ID, CreatedAt, UpdatedAt, DeletedAt等通用字段。
	ContentType  ContentType      `gorm:"type:varchar(32);not null;comment:内容类型" json:"content_type"` // 待审核内容的类型。
	ContentID    uint64           `gorm:"not null;index;comment:内容ID" json:"content_id"`              // 待审核内容的唯一标识符，索引字段。
	Content      string           `gorm:"type:text;comment:内容" json:"content"`                        // 待审核的实际内容。
	UserID       uint64           `gorm:"not null;index;comment:用户ID" json:"user_id"`                 // 提交内容的用户ID，索引字段。
	Status       ModerationStatus `gorm:"default:0;comment:状态" json:"status"`                         // 审核状态，默认为待审核。
	AIScore      float64          `gorm:"type:decimal(5,4);comment:AI评分" json:"ai_score"`             // AI审核给出的评分，例如风险程度。
	AITags       []string         `gorm:"type:json;serializer:json;comment:AI标签" json:"ai_tags"`      // AI审核给出的标签列表。
	RejectReason string           `gorm:"type:varchar(255);comment:拒绝原因" json:"reject_reason"`        // 人工审核拒绝的原因。
	ModeratorID  uint64           `gorm:"comment:审核人ID" json:"moderator_id"`                          // 执行人工审核的管理员ID。
	ModeratedAt  *time.Time       `gorm:"comment:审核时间" json:"moderated_at"`                           // 人工审核完成的时间。
}

// SensitiveWord 实体代表一个敏感词。
// 敏感词用于过滤和检测内容中的违规信息。
type SensitiveWord struct {
	gorm.Model        // 嵌入gorm.Model。
	Word       string `gorm:"type:varchar(64);uniqueIndex;not null;comment:敏感词" json:"word"` // 敏感词，唯一索引，不允许为空。
	Category   string `gorm:"type:varchar(32);not null;comment:分类" json:"category"`          // 敏感词的分类，例如“政治”、“色情”。
	Level      int8   `gorm:"default:1;comment:等级" json:"level"`                             // 敏感词的敏感等级，默认为1。
	Enabled    bool   `gorm:"default:true;comment:是否启用" json:"enabled"`                      // 敏感词是否启用，默认为启用。
}

// NewModerationRecord 创建并返回一个新的 ModerationRecord 实体实例。
// contentType: 内容类型。
// contentID: 内容ID。
// content: 内容字符串。
// userID: 提交内容的用户ID。
func NewModerationRecord(contentType ContentType, contentID uint64, content string, userID uint64) *ModerationRecord {
	return &ModerationRecord{
		ContentType: contentType,
		ContentID:   contentID,
		Content:     content,
		UserID:      userID,
		Status:      ModerationStatusPending, // 新创建的记录默认为待审核状态。
		AITags:      []string{},              // 初始化AI标签列表。
	}
}

// SetAIResult 设置AI审核结果，并根据AI评分自动进行初步审核。
// score: AI给出的风险评分。
// tags: AI识别出的内容标签。
func (m *ModerationRecord) SetAIResult(score float64, tags []string) {
	m.AIScore = score
	m.AITags = tags

	// 自动审核逻辑：
	// 如果AI评分很低（例如，小于0.3），则自动批准。
	if score < 0.3 {
		m.AutoApprove()
	} else if score > 0.8 {
		// 如果AI评分很高（例如，大于0.8），则自动拒绝。
		m.AutoReject("AI检测到违规内容")
	}
}

// Approve 批准审核记录，更新状态为“通过”，并记录审核人和审核时间。
// moderatorID: 执行审核的管理员ID。
func (m *ModerationRecord) Approve(moderatorID uint64) {
	m.Status = ModerationStatusApproved // 状态更新为“通过”。
	m.ModeratorID = moderatorID         // 记录审核人ID。
	now := time.Now()
	m.ModeratedAt = &now // 记录审核时间。
}

// Reject 拒绝审核记录，更新状态为“拒绝”，并记录审核人、拒绝原因和审核时间。
// moderatorID: 执行审核的管理员ID。
// reason: 拒绝的原因。
func (m *ModerationRecord) Reject(moderatorID uint64, reason string) {
	m.Status = ModerationStatusRejected // 状态更新为“拒绝”。
	m.ModeratorID = moderatorID         // 记录审核人ID。
	m.RejectReason = reason             // 记录拒绝原因。
	now := time.Now()
	m.ModeratedAt = &now // 记录审核时间。
}

// AutoApprove 自动批准审核记录。
func (m *ModerationRecord) AutoApprove() {
	m.Status = ModerationStatusApproved // 状态更新为“通过”。
	now := time.Now()
	m.ModeratedAt = &now // 记录审核时间。
}

// AutoReject 自动拒绝审核记录。
// reason: 拒绝的原因。
func (m *ModerationRecord) AutoReject(reason string) {
	m.Status = ModerationStatusRejected // 状态更新为“拒绝”。
	m.RejectReason = reason             // 记录拒绝原因。
	now := time.Now()
	m.ModeratedAt = &now // 记录审核时间。
}

// NewSensitiveWord 创建并返回一个新的 SensitiveWord 实体实例。
// word: 敏感词字符串。
// category: 敏感词分类。
// level: 敏感等级。
func NewSensitiveWord(word, category string, level int8) *SensitiveWord {
	return &SensitiveWord{
		Word:     word,
		Category: category,
		Level:    level,
		Enabled:  true, // 默认启用。
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
