package domain

import "time"

const (
	ModerationRecordCreatedEventType = "contentmoderation.record.created"
	ModerationRecordUpdatedEventType = "contentmoderation.record.updated"
	ModerationRecordDeletedEventType = "contentmoderation.record.deleted"

	SensitiveWordCreatedEventType = "contentmoderation.word.created"
	SensitiveWordUpdatedEventType = "contentmoderation.word.updated"
	SensitiveWordDeletedEventType = "contentmoderation.word.deleted"
)

// ModerationRecordCreatedEvent 审核记录创建事件。
type ModerationRecordCreatedEvent struct {
	RecordID    uint64           `json:"record_id"`
	ContentType ContentType      `json:"content_type"`
	ContentID   uint64           `json:"content_id"`
	UserID      uint64           `json:"user_id"`
	Status      ModerationStatus `json:"status"`
	Timestamp   time.Time        `json:"timestamp"`
}

// ModerationRecordUpdatedEvent 审核记录更新事件（人工审核/自动审核）。
type ModerationRecordUpdatedEvent struct {
	RecordID    uint64           `json:"record_id"`
	Status      ModerationStatus `json:"status"`
	ModeratorID uint64           `json:"moderator_id"`
	Timestamp   time.Time        `json:"timestamp"`
}

// ModerationRecordDeletedEvent 审核记录删除事件。
type ModerationRecordDeletedEvent struct {
	RecordID  uint64    `json:"record_id"`
	Timestamp time.Time `json:"timestamp"`
}

// SensitiveWordCreatedEvent 敏感词创建事件。
type SensitiveWordCreatedEvent struct {
	WordID    uint64    `json:"word_id"`
	Word      string    `json:"word"`
	Timestamp time.Time `json:"timestamp"`
}

// SensitiveWordUpdatedEvent 敏感词更新事件。
type SensitiveWordUpdatedEvent struct {
	WordID    uint64    `json:"word_id"`
	Word      string    `json:"word"`
	Timestamp time.Time `json:"timestamp"`
}

// SensitiveWordDeletedEvent 敏感词删除事件。
type SensitiveWordDeletedEvent struct {
	WordID    uint64    `json:"word_id"`
	Timestamp time.Time `json:"timestamp"`
}
