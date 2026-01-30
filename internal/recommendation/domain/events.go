package domain

import "time"

// RecommendationGeneratedEvent 推荐生成事件。
type RecommendationGeneratedEvent struct {
	UserID    uint64    `json:"user_id"`
	Scope     string    `json:"scope"`
	ItemCount int       `json:"item_count"`
	Timestamp time.Time `json:"timestamp"`
}

// UserPreferenceUpdatedEvent 用户偏好更新事件。
type UserPreferenceUpdatedEvent struct {
	UserID    uint64    `json:"user_id"`
	Action    string    `json:"action"`
	Timestamp time.Time `json:"timestamp"`
}
