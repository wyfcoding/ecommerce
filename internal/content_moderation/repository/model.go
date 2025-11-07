package repository

import (
	"time"

	"gorm.io/gorm"
)

// ModerationResult represents a record of a content moderation check.
type ModerationResult struct {
	gorm.Model
	ContentID   string    `gorm:"index;not null;comment:内容ID (e.g., review_id, post_id)" json:"contentId"`
	ContentType string    `gorm:"size:50;not null;comment:内容类型 (e.g., review, comment)" json:"contentType"`
	UserID      string    `gorm:"index;comment:用户ID" json:"userId"`
	TextContent string    `gorm:"type:text;comment:文本内容" json:"textContent"`
	ImageURL    string    `gorm:"type:text;comment:图片URL" json:"imageUrl"`
	IsSafe      bool      `gorm:"not null;comment:是否安全" json:"isSafe"`
	Labels      string    `gorm:"type:json;comment:识别出的标签 (JSON数组)" json:"labels"` // Stored as JSON array
	Confidence  float64   `gorm:"comment:置信度" json:"confidence"`
	Decision    string    `gorm:"size:50;not null;comment:决策 (ALLOW, REVIEW, REJECT)" json:"decision"`
	ModeratedAt time.Time `gorm:"not null;comment:审核时间" json:"moderatedAt"`
}

// TableName specifies the table name for ModerationResult.
func (ModerationResult) TableName() string {
	return "content_moderation_results"
}
