package domain

import "time"

// ContentModeratedEvent 内容审核完成事件。
type ContentModeratedEvent struct {
	ModerationID uint64    `json:"moderation_id"`
	SourceType   string    `json:"source_type"`
	SourceID     uint64    `json:"source_id"`
	Status       string    `json:"status"`
	Timestamp    time.Time `json:"timestamp"`
}

// ModerationPolicyUpdatedEvent 审核策略更新事件。
type ModerationPolicyUpdatedEvent struct {
	PolicyID  uint64    `json:"policy_id"`
	Action    string    `json:"action"`
	Timestamp time.Time `json:"timestamp"`
}
