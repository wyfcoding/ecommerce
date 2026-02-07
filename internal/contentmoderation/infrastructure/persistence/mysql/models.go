package mysql

import (
	"time"

	"github.com/wyfcoding/ecommerce/internal/contentmoderation/domain"
	"gorm.io/gorm"
)

// ModerationRecordModel 审核记录写模型。
type ModerationRecordModel struct {
	gorm.Model
	ContentType  domain.ContentType      `gorm:"column:content_type;type:varchar(16);not null;index;comment:内容类型"`
	ContentID    uint64                  `gorm:"column:content_id;not null;index;comment:内容ID"`
	Content      string                  `gorm:"column:content;type:longtext;not null;comment:内容"`
	UserID       uint64                  `gorm:"column:user_id;not null;index;comment:用户ID"`
	Status       domain.ModerationStatus `gorm:"column:status;type:tinyint;not null;index;comment:审核状态"`
	AIScore      float64                 `gorm:"column:ai_score;type:double;comment:AI评分"`
	AITags       []string                `gorm:"column:ai_tags;type:json;serializer:json;comment:AI标签"`
	RejectReason string                  `gorm:"column:reject_reason;type:varchar(255);comment:拒绝原因"`
	ModeratorID  uint64                  `gorm:"column:moderator_id;index;comment:审核员ID"`
	ModeratedAt  *time.Time              `gorm:"column:moderated_at;comment:审核时间"`
}

// SensitiveWordModel 敏感词写模型。
type SensitiveWordModel struct {
	gorm.Model
	Word     string `gorm:"column:word;type:varchar(128);uniqueIndex;not null;comment:敏感词"`
	Category string `gorm:"column:category;type:varchar(64);comment:分类"`
	Level    int8   `gorm:"column:level;type:tinyint;comment:等级"`
	Enabled  bool   `gorm:"column:enabled;default:true;comment:是否启用"`
}

func (ModerationRecordModel) TableName() string { return "moderation_records" }
func (SensitiveWordModel) TableName() string    { return "sensitive_words" }

func toModerationRecordModel(record *domain.ModerationRecord) *ModerationRecordModel {
	if record == nil {
		return nil
	}
	return &ModerationRecordModel{
		Model: gorm.Model{
			ID:        record.ID,
			CreatedAt: record.CreatedAt,
			UpdatedAt: record.UpdatedAt,
		},
		ContentType:  record.ContentType,
		ContentID:    record.ContentID,
		Content:      record.Content,
		UserID:       record.UserID,
		Status:       record.Status,
		AIScore:      record.AIScore,
		AITags:       record.AITags,
		RejectReason: record.RejectReason,
		ModeratorID:  record.ModeratorID,
		ModeratedAt:  record.ModeratedAt,
	}
}

func toModerationRecord(model *ModerationRecordModel) *domain.ModerationRecord {
	if model == nil {
		return nil
	}
	return &domain.ModerationRecord{
		ID:           model.ID,
		CreatedAt:    model.CreatedAt,
		UpdatedAt:    model.UpdatedAt,
		ContentType:  model.ContentType,
		ContentID:    model.ContentID,
		Content:      model.Content,
		UserID:       model.UserID,
		Status:       model.Status,
		AIScore:      model.AIScore,
		AITags:       model.AITags,
		RejectReason: model.RejectReason,
		ModeratorID:  model.ModeratorID,
		ModeratedAt:  model.ModeratedAt,
	}
}

func toSensitiveWordModel(word *domain.SensitiveWord) *SensitiveWordModel {
	if word == nil {
		return nil
	}
	return &SensitiveWordModel{
		Model: gorm.Model{
			ID:        word.ID,
			CreatedAt: word.CreatedAt,
			UpdatedAt: word.UpdatedAt,
		},
		Word:     word.Word,
		Category: word.Category,
		Level:    word.Level,
		Enabled:  word.Enabled,
	}
}

func toSensitiveWord(model *SensitiveWordModel) *domain.SensitiveWord {
	if model == nil {
		return nil
	}
	return &domain.SensitiveWord{
		ID:        model.ID,
		CreatedAt: model.CreatedAt,
		UpdatedAt: model.UpdatedAt,
		Word:      model.Word,
		Category:  model.Category,
		Level:     model.Level,
		Enabled:   model.Enabled,
	}
}
