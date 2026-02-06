package domain

import "time"

const (
	RecommendationChangedEventType = "recommendation.changed"
	RecommendationDeletedEventType = "recommendation.deleted"
	UserPreferenceUpdatedEventType = "recommendation.preference.updated"
	UserBehaviorRecordedEventType  = "recommendation.behavior.recorded"
)

// RecommendationChangedEvent 推荐列表发生变化事件。
type RecommendationChangedEvent struct {
	UserID             uint64             `json:"user_id"`
	RecommendationType RecommendationType `json:"recommendation_type"`
	Timestamp          time.Time          `json:"timestamp"`
}

// RecommendationDeletedEvent 推荐删除事件。
type RecommendationDeletedEvent struct {
	UserID             uint64             `json:"user_id"`
	RecommendationType RecommendationType `json:"recommendation_type"`
	Timestamp          time.Time          `json:"timestamp"`
}

// UserPreferenceUpdatedEvent 用户偏好更新事件。
type UserPreferenceUpdatedEvent struct {
	UserID    uint64    `json:"user_id"`
	Timestamp time.Time `json:"timestamp"`
}

// UserBehaviorRecordedEvent 用户行为记录事件。
type UserBehaviorRecordedEvent struct {
	UserID    uint64    `json:"user_id"`
	ProductID uint64    `json:"product_id"`
	Action    string    `json:"action"`
	Timestamp time.Time `json:"timestamp"`
}
